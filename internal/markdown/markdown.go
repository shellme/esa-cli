package markdown

import (
	"fmt"
	"strings"

	"github.com/shellme/esa-cli/pkg/types"
)

// MarkdownコンテンツにメタデータのFrontMatterを追加
func GenerateContent(post types.Post) string {
	var content strings.Builder

	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("title: %s\n", post.Name))
	if post.Category != "" {
		content.WriteString(fmt.Sprintf("category: %s\n", post.Category))
	}
	if len(post.Tags) > 0 {
		content.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(post.Tags, ", ")))
	}
	content.WriteString(fmt.Sprintf("wip: %t\n", post.WIP))
	content.WriteString("---\n\n")
	content.WriteString(post.BodyMd)

	return content.String()
}

// Markdownファイルからメタデータと本文を分離
func ParseContent(content string) (types.Post, error) {
	lines := strings.Split(content, "\n")

	var post types.Post
	inFrontMatter := false
	bodyStartIndex := 0

	// Front Matterを解析
	for i, line := range lines {
		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			} else {
				bodyStartIndex = i + 1
				if bodyStartIndex < len(lines) && lines[bodyStartIndex] == "" {
					bodyStartIndex++
				}
				break
			}
		}

		if inFrontMatter {
			if strings.HasPrefix(line, "title: ") {
				post.Name = strings.TrimPrefix(line, "title: ")
			} else if strings.HasPrefix(line, "category: ") {
				post.Category = strings.TrimPrefix(line, "category: ")
			} else if strings.HasPrefix(line, "tags: ") {
				tagsStr := strings.TrimPrefix(line, "tags: ")
				tagsStr = strings.Trim(tagsStr, "[]")
				if tagsStr != "" {
					post.Tags = strings.Split(tagsStr, ", ")
					// タグの前後の空白を除去
					for i, tag := range post.Tags {
						post.Tags[i] = strings.TrimSpace(tag)
					}
				}
			} else if strings.HasPrefix(line, "wip: ") {
				wipStr := strings.TrimPrefix(line, "wip: ")
				post.WIP = wipStr == "true"
			}
		}
	}

	// 本文を結合
	if bodyStartIndex < len(lines) {
		post.BodyMd = strings.Join(lines[bodyStartIndex:], "\n")
	}

	post.Message = "Updated via esa-cli"

	return post, nil
}
