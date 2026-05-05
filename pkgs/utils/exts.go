package utils

var (
	VideoExts = map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".flv":  true,
		".avi":  true,
		".wmv":  true,
		".ts":   true,
		".rmvb": true,
		".webm": true,
		".mpg":  true,
		".m2ts": true,
	}

	SubtitleExts = map[string]bool{
		".ass": true,
		".srt": true,
		".ssa": true,
		".sub": true,
	}

	ImageExts = map[string]bool{
		".png": true,
		".jpg": true,
	}

	NfoExts = map[string]bool{
		".nfo": true,
	}
)

func IsExtendedVideoExt(ext string) bool {
	if ext == ".strm" {
		return true
	}
	return VideoExts[ext]
}
