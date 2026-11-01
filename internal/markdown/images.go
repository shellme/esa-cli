package markdown

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	imagePattern  = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	schemePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)
	fencePattern  = regexp.MustCompile("^ {0,3}(```|~~~)")
	imageExts     = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true}
)

// isLocalImage URLではなく画像の拡張子を持つパスかどうかを判定する
func isLocalImage(path string) bool {
	if schemePattern.MatchString(path) || strings.HasPrefix(path, "//") {
		return false
	}
	return imageExts[strings.ToLower(filepath.Ext(path))]
}

// ReplaceLocalImages コードを除く本文のローカル画像パスをreplaceの結果で置換する
func ReplaceLocalImages(body string, replace func(path string) string) string {
	var result strings.Builder
	var fence string
	for _, line := range strings.SplitAfter(body, "\n") {
		if match := fencePattern.FindStringSubmatch(line); match != nil {
			if fence == "" {
				fence = match[1]
			} else if fence == match[1] {
				fence = ""
			}
			result.WriteString(line)
			continue
		}
		if fence != "" {
			result.WriteString(line)
			continue
		}
		result.WriteString(replaceInLine(line, replace))
	}
	return result.String()
}

func replaceInLine(line string, replace func(string) string) string {
	var result strings.Builder
	last := 0
	for _, match := range imagePattern.FindAllStringSubmatchIndex(line, -1) {
		path := line[match[2]:match[3]]
		// バッククォートが奇数個ならインラインコードの中
		if strings.Count(line[:match[0]], "`")%2 == 1 || !isLocalImage(path) {
			continue
		}
		result.WriteString(line[last:match[2]])
		result.WriteString(replace(path))
		last = match[3]
	}
	result.WriteString(line[last:])
	return result.String()
}

// LocalImages 本文に含まれるローカル画像パスを重複なしで返す
func LocalImages(body string) []string {
	var paths []string
	seen := map[string]bool{}
	ReplaceLocalImages(body, func(path string) string {
		if !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
		return path
	})
	return paths
}

// ImageFilePath 本文中のパスをファイルシステム上のパスに変換する
func ImageFilePath(articlePath, imagePath string) string {
	if decoded, err := url.PathUnescape(imagePath); err == nil {
		imagePath = decoded
	}
	if filepath.IsAbs(imagePath) {
		return imagePath
	}
	return filepath.Join(filepath.Dir(articlePath), imagePath)
}
