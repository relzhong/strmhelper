package openlist2strm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/drivers/openlist"
	"github.com/go-resty/resty/v2"
	"github.com/relzhong/strmhelper/pkgs/config"
	"github.com/relzhong/strmhelper/pkgs/utils"
	"log/slog"
	"net/url"
	"path"
)

type OpenList2Strm struct {
	ctx               context.Context
	client            *openListClient
	mode              OpenList2StrmMode
	publicURL         string
	sourceDir         string
	targetDir         string
	flattenMode       bool
	downloadExts      map[string]bool
	processFileExts   map[string]bool
	overwrite         bool
	syncServer        bool
	syncIgnorePattern *regexp.Regexp
	strmProtection    *StrmProtectionManager
	checkModTime      bool
	syncBack          bool
	syncState         *SyncStateManager
	taskID            string

	bdmvCollections   map[string][]bdmvFileItem
	bdmvLargestFiles  map[string]openlist.ObjResp
	processedLocalMap map[string]bool
}

type bdmvFileItem struct {
	path openlist.ObjResp
	size int64
}

type SmartProtectionConfig struct {
	Enabled    bool
	Threshold  int
	GraceScans int
}

func NewOpenList2Strm(
	ctx context.Context,
	id string, url string, username string, password string, token string,
	publicURL string, sourceDir string, targetDir string, flattenMode bool,
	subtitle bool, image bool, nfo bool, mode string, overwrite bool,
	otherExt string, maxWorkers int, maxDownloaders int, waitTime float64,
	syncServer bool, syncIgnore string, smartProtection *SmartProtectionConfig, checkModTime bool,
	syncBack bool,
) (*OpenList2Strm, error) {

	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	url = strings.TrimRight(url, "/")

	client := &openListClient{
		Addition: openlist.Addition{
			Address:  url,
			Username: username,
			Password: password,
			Token:    token,
		},
		resty: resty.New(),
	}
	client.resty.SetHeader("User-Agent", "StrmHelper/"+config.Settings.AppVersion)

	err := client.Init()
	if err != nil {
		return nil, err
	}

	if publicURL != "" && !strings.HasPrefix(publicURL, "http") {
		publicURL = "https://" + publicURL
	}
	publicURL = strings.TrimRight(publicURL, "/")

	downloadExts := make(map[string]bool)
	if subtitle {
		for k := range utils.SubtitleExts {
			downloadExts[k] = true
		}
	}
	if image {
		for k := range utils.ImageExts {
			downloadExts[k] = true
		}
	}
	if nfo {
		for k := range utils.NfoExts {
			downloadExts[k] = true
		}
	}
	if otherExt != "" {
		for _, ext := range strings.Split(strings.ToLower(otherExt), ",") {
			downloadExts[ext] = true
		}
	}

	processFileExts := make(map[string]bool)
	for k := range utils.VideoExts {
		processFileExts[k] = true
	}
	for k := range downloadExts {
		processFileExts[k] = true
	}

	var syncIgnoreRe *regexp.Regexp
	if syncIgnore != "" {
		syncIgnoreRe = regexp.MustCompile(syncIgnore)
	}

	var strmProt *StrmProtectionManager
	graceScans := 0
	if smartProtection != nil && smartProtection.Enabled {
		strmProt = NewStrmProtectionManager(targetDir, id, smartProtection.Threshold, smartProtection.GraceScans)
		graceScans = smartProtection.GraceScans
		slog.Info(fmt.Sprintf(".strm 保护已启用：阈值=%d，宽限期=%d", smartProtection.Threshold, smartProtection.GraceScans))
	}

	return &OpenList2Strm{
		ctx:               ctx,
		client:            client,
		mode:              ModeFromStr(mode),
		publicURL:         publicURL,
		sourceDir:         sourceDir,
		targetDir:         targetDir,
		flattenMode:       flattenMode,
		downloadExts:      downloadExts,
		processFileExts:   processFileExts,
		overwrite:         overwrite,
		syncServer:        syncServer,
		syncIgnorePattern: syncIgnoreRe,
		strmProtection:    strmProt,
		checkModTime:      checkModTime,
		syncBack:          syncBack,
		syncState:         NewSyncStateManager(id, targetDir, graceScans),
		taskID:            id,

		bdmvCollections:   make(map[string][]bdmvFileItem),
		bdmvLargestFiles:  make(map[string]openlist.ObjResp),
		processedLocalMap: make(map[string]bool),
	}, nil
}

func (a *OpenList2Strm) Run() error {
	select {
	case <-a.ctx.Done():
		return a.ctx.Err()
	default:
	}

	a.bdmvCollections = make(map[string][]bdmvFileItem)
	a.bdmvLargestFiles = make(map[string]openlist.ObjResp)
	a.processedLocalMap = make(map[string]bool)

	// Safety Check: If targetDir is empty but DB has synced files, abort (likely unmounted drive)
	if a.syncBack {
		localEntries, _ := os.ReadDir(a.targetDir)
		if len(localEntries) == 0 && a.syncState.GetSyncedCount() > 0 {
			return fmt.Errorf("Safety Check Failed: targetDir is empty but database contains %d synced files. Is the drive unmounted?", a.syncState.GetSyncedCount())
		}
	}

	filter := func(p openlist.ObjResp, fullPath string) bool {
		if p.IsDir {
			slog.Debug("Filter: skipping directory object", "path", fullPath)
			return false
		}
		if strings.Contains(fullPath, "@eaDir") || strings.Contains(fullPath, "Thumbs.db") || strings.Contains(fullPath, ".DS_Store") {
			return false
		}
		if strings.Contains(fullPath, "/BDMV/") && !a.isBdmvFile(p, fullPath) {
			return false
		}

		suffix := strings.ToLower(filepath.Ext(p.Name))
		if !a.processFileExts[suffix] {
			return false
		}

		if a.isBdmvFile(p, fullPath) {
			a.collectBdmvFile(p, fullPath)
			return false
		}

		localPath := a.getLocalPath(p, fullPath)

		if a.syncBack {
			if _, err := os.Stat(localPath); os.IsNotExist(err) {
				if a.syncState.IsKnown(localPath) {
					// Local delete detected
					if a.syncState.HandleMissingLocal(localPath, fullPath) {
						slog.Warn("Syncing back delete to remote", "remote", fullPath)
						err := a.client.Remove(fullPath)
						if err != nil {
							slog.Error("Failed to remove remote file", "path", fullPath, "error", err)
						} else {
							a.syncState.Unmark(localPath)
						}
					}
					return false // Skip processing this file (it's either pending delete or just deleted)
				}
			} else {
				// File exists locally, clear any pending remote deletes
				a.syncState.ClearPending(fullPath)
			}
		}

		a.processedLocalMap[localPath] = true
		// Mark as synced if not already
		a.syncState.MarkSynced(localPath, fullPath)

		if !a.overwrite {
			if stat, err := os.Stat(localPath); err == nil {
				if a.downloadExts[suffix] {
					modTime := float64(stat.ModTime().UnixNano()) / 1e9
					pModTime := float64(p.Modified.UnixNano()) / 1e9
					if modTime < pModTime {
						return true
					}
					if stat.Size() < p.Size {
						return true
					}
				}
				return false
			}
		}
		return true
	}

	err := a.IterPath(a.sourceDir, filter, a.fileProcesser)
	if err != nil {
		return fmt.Errorf("IterPath Error: %v", err)
	}

	a.finalizeBdmvCollections()

	for fullPath, largestFile := range a.bdmvLargestFiles {
		select {
		case <-a.ctx.Done():
			return a.ctx.Err()
		default:
		}
		a.fileProcesser(largestFile, fullPath)
		a.processedLocalMap[a.getLocalPath(largestFile, fullPath)] = true
	}

	if a.syncServer {
		a.cleanupLocalFiles()
	}

	slog.Info("OpenList2Strm 处理完成")
	return nil
}

func (a *OpenList2Strm) fileProcesser(p openlist.ObjResp, fullPath string) {
	localPath := a.getLocalPath(p, fullPath)
	var content string

	switch a.mode {
	case OpenListURL:
		content = a.client.getDownloadURL(p, fullPath)
		if a.publicURL != "" {
			parsedContent, err := url.Parse(content)
			if err == nil {
				parsedPublic, err := url.Parse(a.publicURL)
				if err == nil {
					parsedContent.Scheme = parsedPublic.Scheme
					parsedContent.Host = parsedPublic.Host
					// If publicURL has a path prefix, prepend it
					if parsedPublic.Path != "" && parsedPublic.Path != "/" {
						parsedContent.Path = path.Join(parsedPublic.Path, parsedContent.Path)
					}
					content = parsedContent.String()
				}
			}
		}
	case RawURL:
		content = a.client.getRawURL(p, fullPath)
	case OpenListPath:
		content = fullPath
	}

	if content == "" {
		return
	}

	os.MkdirAll(filepath.Dir(localPath), 0755)

	if strings.HasSuffix(localPath, ".strm") {
		os.WriteFile(localPath, []byte(content), 0644)
		slog.Info(fmt.Sprintf("%s 创建成功", filepath.Base(localPath)))
	} else {
		downloadURL := a.client.getDownloadURL(p, fullPath)
		err := a.client.DownloadFile(downloadURL, localPath)
		if err == nil {
			slog.Info(fmt.Sprintf("%s 下载成功", filepath.Base(localPath)))
			a.syncState.MarkSynced(localPath, downloadURL) // Update with download URL if needed, but path is better
		}
	}
}

func (a *OpenList2Strm) getLocalPath(p openlist.ObjResp, fullPath string) string {
	if a.isBdmvFile(p, fullPath) {
		bdmvRoot := a.getBdmvRootDir(fullPath)
		if bdmvRoot != "" && a.shouldProcessBdmvFile(p, fullPath) {
			movieTitle := filepath.Base(bdmvRoot)
			if a.flattenMode {
				return filepath.Join(a.targetDir, movieTitle+".strm")
			}
			rel := strings.TrimPrefix(bdmvRoot, a.sourceDir)
			rel = strings.TrimPrefix(rel, "/")
			return filepath.Join(a.targetDir, rel, movieTitle+".strm")
		}
	}

	var localPath string
	if a.flattenMode {
		localPath = filepath.Join(a.targetDir, p.Name)
	} else {
		rel := strings.TrimPrefix(fullPath, a.sourceDir)
		rel = strings.TrimPrefix(rel, "/")
		localPath = filepath.Join(a.targetDir, rel)
	}

	suffix := strings.ToLower(filepath.Ext(p.Name))
	if utils.VideoExts[suffix] {
		ext := filepath.Ext(localPath)
		localPath = localPath[:len(localPath)-len(ext)] + ".strm"
	}

	return localPath
}

func (a *OpenList2Strm) cleanupLocalFiles() {
	var allLocalFiles []string
	if a.flattenMode {
		entries, _ := os.ReadDir(a.targetDir)
		for _, e := range entries {
			if !e.IsDir() {
				allLocalFiles = append(allLocalFiles, filepath.Join(a.targetDir, e.Name()))
			}
		}
	} else {
		filepath.Walk(a.targetDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				allLocalFiles = append(allLocalFiles, path)
			}
			return nil
		})
	}

	filesToDelete := make(map[string]bool)
	for _, f := range allLocalFiles {
		if !a.processedLocalMap[f] {
			filesToDelete[f] = true
		}
	}

	strmPresent := make(map[string]bool)
	if a.strmProtection != nil {
		for f := range a.processedLocalMap {
			if strings.HasSuffix(f, ".strm") {
				strmPresent[f] = true
			}
		}
	}

	if len(filesToDelete) == 0 {
		if a.strmProtection != nil {
			a.strmProtection.Process(make(map[string]bool), strmPresent)
			a.strmProtection.Save()
		}
		return
	}

	strmToDelete := make(map[string]bool)
	otherFiles := make(map[string]bool)
	for f := range filesToDelete {
		if strings.HasSuffix(f, ".strm") {
			strmToDelete[f] = true
		} else {
			otherFiles[f] = true
		}
	}

	if a.strmProtection != nil {
		strmToDelete = a.strmProtection.Process(strmToDelete, strmPresent)
		a.strmProtection.Save()
	}

	finalDelete := make(map[string]bool)
	for f := range strmToDelete {
		finalDelete[f] = true
	}
	for f := range otherFiles {
		finalDelete[f] = true
	}

	for f := range finalDelete {
		if a.syncIgnorePattern != nil && a.syncIgnorePattern.MatchString(filepath.Base(f)) {
			continue
		}
		os.Remove(f)
		slog.Info(fmt.Sprintf("删除文件：%s", f))
		a.syncState.Unmark(f)

		parent := filepath.Dir(f)
		for parent != a.targetDir {
			entries, _ := os.ReadDir(parent)
			if len(entries) == 0 {
				os.Remove(parent)
				parent = filepath.Dir(parent)
			} else {
				break
			}
		}
	}
}

func (a *OpenList2Strm) isBdmvFile(p openlist.ObjResp, fullPath string) bool {
	return strings.Contains(fullPath, "/BDMV/STREAM/") && strings.ToLower(filepath.Ext(p.Name)) == ".m2ts"
}

func (a *OpenList2Strm) getBdmvRootDir(fullPath string) string {
	idx := strings.Index(fullPath, "/BDMV/")
	if idx != -1 {
		return fullPath[:idx]
	}
	return ""
}

func (a *OpenList2Strm) collectBdmvFile(p openlist.ObjResp, fullPath string) {
	root := a.getBdmvRootDir(fullPath)
	if root != "" {
		a.bdmvCollections[root] = append(a.bdmvCollections[root], bdmvFileItem{p, p.Size})
	}
}

func (a *OpenList2Strm) finalizeBdmvCollections() {
	for root, items := range a.bdmvCollections {
		if len(items) == 0 {
			continue
		}
		var maxItem bdmvFileItem
		for i, item := range items {
			if i == 0 || item.size > maxItem.size {
				maxItem = item
			}
		}
		a.bdmvLargestFiles[root] = maxItem.path
	}
}

func (a *OpenList2Strm) shouldProcessBdmvFile(p openlist.ObjResp, fullPath string) bool {
	root := a.getBdmvRootDir(fullPath)
	if root == "" {
		return false
	}
	if largest, ok := a.bdmvLargestFiles[root]; ok {
		return largest.Name == p.Name && largest.Size == p.Size
	}
	return false
}

func (a *OpenList2Strm) IterPath(dirPath string, filter func(openlist.ObjResp, string) bool, onFile func(openlist.ObjResp, string)) error {
	var recurse func(path string) error
	recurse = func(p string) error {
		select {
		case <-a.ctx.Done():
			return a.ctx.Err()
		default:
		}
		slog.Debug("Listing directory", "path", p)
		items, err := a.client.List(p)
		if err != nil {
			return err
		}

		for _, item := range items {
			select {
			case <-a.ctx.Done():
				return a.ctx.Err()
			default:
			}
			fullPath := p + "/" + item.Name
			if item.IsDir {
				if a.checkModTime {
					localDir := a.getLocalDirPath(fullPath)
					info, err := os.Stat(localDir)
					if err == nil {
						// Compare times
						if !item.Modified.After(info.ModTime()) {
							slog.Info("Skipping unchanged directory", "path", fullPath, "remoteMod", item.Modified, "localMod", info.ModTime())
							a.markLocalFilesAsProcessed(localDir)
							continue
						} else {
							slog.Debug("Directory changed, scanning", "path", fullPath, "remoteMod", item.Modified, "localMod", info.ModTime())
						}
					} else {
						slog.Debug("Local directory not found, scanning", "path", fullPath)
					}
				}
				err := recurse(fullPath)
				if err != nil {
					return err
				}
				if a.checkModTime {
					// After successful recursion, sync the local directory time with the remote one
					// to handle potential clock skew in future runs.
					localDir := a.getLocalDirPath(fullPath)
					os.Chtimes(localDir, item.Modified, item.Modified)
				}
			}

			if filter == nil || filter(item, fullPath) {
				slog.Debug("Processing item", "name", item.Name, "path", fullPath)
				if onFile != nil {
					onFile(item, fullPath)
				}
			}
		}
		return nil
	}

	if a.checkModTime {
		remoteInfo, err := a.client.GetFileInfo(dirPath)
		if err == nil {
			localDir := a.targetDir
			info, err := os.Stat(localDir)
			if err == nil {
				if !remoteInfo.Modified.After(info.ModTime()) {
					slog.Info("Skipping unchanged root directory", "path", dirPath, "remoteMod", remoteInfo.Modified, "localMod", info.ModTime())
					a.markLocalFilesAsProcessed(localDir)
					return nil
				}
			}
			// Root sync will happen at the end of recurse if we don't return here
			defer func() {
				os.Chtimes(localDir, remoteInfo.Modified, remoteInfo.Modified)
			}()
		}
	}

	if err := recurse(dirPath); err != nil {
		return err
	}

	return nil
}

// openListClient is a simplified implementation that uses official OpenList v4 types
type openListClient struct {
	openlist.Addition
	resty    *resty.Client
	basePath string
}

func (c *openListClient) Init() error {
	c.Addition.Address = strings.TrimSuffix(c.Addition.Address, "/")
	
	var resp struct {
		Code int             `json:"code"`
		Data openlist.MeResp `json:"data"`
	}
	_, err := c.resty.R().
		SetHeader("Authorization", c.Token).
		SetResult(&resp).
		Get(c.Address + "/api/me")
	
	if err == nil && resp.Code == 200 {
		c.basePath = resp.Data.BasePath
	} else {
		token, err := c.Login()
		if err != nil {
			return err
		}
		c.Token = token
		c.resty.R().
			SetHeader("Authorization", c.Token).
			SetResult(&resp).
			Get(c.Address + "/api/me")
		if resp.Code == 200 {
			c.basePath = resp.Data.BasePath
		}
	}
	return nil
}

func (c *openListClient) Login() (string, error) {
	var resp struct {
		Code int                `json:"code"`
		Data openlist.LoginResp `json:"data"`
	}
	_, err := c.resty.R().
		SetBody(map[string]string{
			"username": c.Username,
			"password": c.Password,
		}).
		SetResult(&resp).
		Post(c.Address + "/api/auth/login")
	
	if err != nil {
		return "", err
	}
	if resp.Code != 200 {
		return "", fmt.Errorf("login failed: %d", resp.Code)
	}
	return resp.Data.Token, nil
}

func (c *openListClient) List(path string) ([]openlist.ObjResp, error) {
	var resp struct {
		Code int                 `json:"code"`
		Data openlist.FsListResp `json:"data"`
	}
	_, err := c.resty.R().
		SetHeader("Authorization", c.Token).
		SetBody(openlist.ListReq{
			Path:     path,
			Password: c.MetaPassword,
		}).
		SetResult(&resp).
		Post(c.Address + "/api/fs/list")
	
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("list failed: %d", resp.Code)
	}
	return resp.Data.Content, nil
}

func (c *openListClient) GetFileInfo(path string) (openlist.ObjResp, error) {
	var resp struct {
		Code int              `json:"code"`
		Data openlist.ObjResp `json:"data"`
	}
	_, err := c.resty.R().
		SetHeader("Authorization", c.Token).
		SetBody(map[string]string{
			"path": path,
		}).
		SetResult(&resp).
		Post(c.Address + "/api/fs/get")

	if err != nil {
		return openlist.ObjResp{}, err
	}
	if resp.Code != 200 {
		return openlist.ObjResp{}, fmt.Errorf("get failed: %d", resp.Code)
	}
	return resp.Data, nil
}

func (c *openListClient) getDownloadURL(p openlist.ObjResp, fullPath string) string {
	u, err := url.Parse(c.Address)
	if err != nil {
		return ""
	}

	// Join /d, basePath and fullPath
	// path.Join handles multiple slashes and cleaning
	fullUrlPath := path.Join("/d", c.basePath, fullPath)
	u.Path = fullUrlPath

	if p.Sign != "" {
		q := u.Query()
		q.Set("sign", p.Sign)
		u.RawQuery = q.Encode()
	}

	return u.String()
}

func (c *openListClient) getRawURL(p openlist.ObjResp, fullPath string) string {
	var resp struct {
		Code int                `json:"code"`
		Data openlist.FsGetResp `json:"data"`
	}
	_, err := c.resty.R().
		SetHeader("Authorization", c.Token).
		SetBody(openlist.FsGetReq{
			Path:     fullPath,
			Password: c.MetaPassword,
		}).
		SetResult(&resp).
		Post(c.Address + "/api/fs/get")
	
	if err == nil && resp.Code == 200 {
		return resp.Data.RawURL
	}
	return ""
}

func (c *openListClient) Remove(fullPath string) error {
	var resp struct {
		Code int `json:"code"`
	}
	_, err := c.resty.R().
		SetHeader("Authorization", c.Token).
		SetBody(map[string]interface{}{
			"dir":   path.Dir(fullPath),
			"names": []string{path.Base(fullPath)},
		}).
		SetResult(&resp).
		Post(c.Address + "/api/fs/remove")

	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("remove failed: %d", resp.Code)
	}
	return nil
}

func (c *openListClient) DownloadFile(url string, localPath string) error {
	_, err := c.resty.R().
		SetOutput(localPath).
		Get(url)
	return err
}
func (a *OpenList2Strm) getLocalDirPath(remotePath string) string {
	relPath := strings.TrimPrefix(remotePath, a.sourceDir)
	return filepath.Join(a.targetDir, relPath)
}
func (a *OpenList2Strm) markLocalFilesAsProcessed(localDir string) {
	filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			a.processedLocalMap[path] = true
		}
		return nil
	})
}
