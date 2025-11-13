package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/shellme/esa-cli/pkg/types"
	"gopkg.in/yaml.v2"
)

// GenerateContent generates markdown content with front matter.
func GenerateContent(fm types.FrontMatter, body string) ([]byte, error) {
	var buf bytes.Buffer

	// ---
	buf.WriteString("---\n")

	// front matter
	out, err := yaml.Marshal(fm)
	if err != nil {
		return nil, err
	}
	buf.Write(out)

	// ---
	buf.WriteString("---\n\n")

	// body
	buf.WriteString(body)

	return buf.Bytes(), nil
}

// ParseContent parses markdown content and separates front matter and body.
// Supports both LF (\n) and CRLF (\r\n) line endings.
func ParseContent(content []byte) (types.FrontMatter, string, error) {
	var parts [][]byte

	// Try LF first (most common)
	parts = bytes.SplitN(content, []byte("---\n"), 3)
	if len(parts) < 3 {
		// Try CRLF as fallback
		parts = bytes.SplitN(content, []byte("---\r\n"), 3)
		if len(parts) < 3 {
			return types.FrontMatter{}, "", fmt.Errorf("failed to parse front matter: front matter delimiter not found. The file may have incorrect line endings (CRLF vs LF). Please ensure the file uses LF line endings or convert it using your editor.")
		}
	}

	var fm types.FrontMatter
	if err := yaml.Unmarshal(parts[1], &fm); err != nil {
		return types.FrontMatter{}, "", fmt.Errorf("failed to parse front matter YAML: %w", err)
	}

	body := strings.TrimSpace(string(parts[2]))

	return fm, body, nil
}
