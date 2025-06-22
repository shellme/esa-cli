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
func ParseContent(content []byte) (types.FrontMatter, string, error) {
	parts := bytes.SplitN(content, []byte("---\n"), 3)
	if len(parts) < 3 {
		return types.FrontMatter{}, "", fmt.Errorf("failed to parse front matter")
	}

	var fm types.FrontMatter
	if err := yaml.Unmarshal(parts[1], &fm); err != nil {
		return types.FrontMatter{}, "", err
	}

	body := strings.TrimSpace(string(parts[2]))

	return fm, body, nil
}
