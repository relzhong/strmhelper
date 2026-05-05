package openlist2strm

import "strings"

type OpenList2StrmMode string

const (
	OpenListURL  OpenList2StrmMode = "OpenListURL"
	RawURL       OpenList2StrmMode = "RawURL"
	OpenListPath OpenList2StrmMode = "OpenListPath"
)

func ModeFromStr(modeStr string) OpenList2StrmMode {
	lower := strings.ToLower(modeStr)
	switch lower {
	case "rawurl":
		return RawURL
	case "openlistpath", "alistpath":
		return OpenListPath
	default:
		return OpenListURL
	}
}
