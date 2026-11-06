package markdown

import (
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	imagePattern  = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	imgTagPattern = regexp.MustCompile(`(?i)<img\s[^>]*?\bsrc\s*=\s*(?:"([^"]+)"|'([^']+)')`)
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
	for _, span := range imagePathSpans(line) {
		path := line[span[1]:span[2]]
		// バッククォートが奇数個ならインラインコードの中
		if span[1] < last || strings.Count(line[:span[0]], "`")%2 == 1 || !isLocalImage(path) {
			continue
		}
		result.WriteString(line[last:span[1]])
		result.WriteString(replace(path))
		last = span[2]
	}
	result.WriteString(line[last:])
	return result.String()
}

// imagePathSpans Markdownの画像記法と<img>タグから、{記述の開始位置, パスの開始位置, パスの終了位置} を出現順に返す
func imagePathSpans(line string) [][3]int {
	var spans [][3]int
	for _, m := range imagePattern.FindAllStringSubmatchIndex(line, -1) {
		spans = append(spans, [3]int{m[0], m[2], m[3]})
	}
	for _, m := range imgTagPattern.FindAllStringSubmatchIndex(line, -1) {
		if m[2] >= 0 {
			spans = append(spans, [3]int{m[0], m[2], m[3]})
		} else {
			spans = append(spans, [3]int{m[0], m[4], m[5]})
		}
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i][0] < spans[j][0] })
	return spans
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
