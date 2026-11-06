package markdown

import (
	"strings"
	"testing"

	"github.com/shellme/esa-cli/pkg/types"
)

func TestGenerateContent(t *testing.T) {
	fm := types.FrontMatter{
		Title:    "Test Article",
		Category: "Test",
		Tags:     []string{"tag1", "tag2"},
		Wip:      false,
	}

	content, err := GenerateContent(fm, "Body content\nLine 2")
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(content) == 0 {
		t.Error("Generated content is empty")
	}

	// Check that content starts with front matter delimiter
	if !startsWith(content, []byte("---\n")) {
		t.Error("Generated content should start with '---\\n'")
	}
}

func TestParseContent_LF(t *testing.T) {
	// Test with LF line endings
	content := []byte("---\ntitle: Test Article\ncategory: Test\ntags:\n  - tag1\n  - tag2\nwip: false\n---\n\nBody content\nLine 2")

	fm, body, err := ParseContent(content)
	if err != nil {
		t.Fatalf("ParseContent failed: %v", err)
	}

	if fm.Title != "Test Article" {
		t.Errorf("Expected title 'Test Article', got '%s'", fm.Title)
	}

	if fm.Category != "Test" {
		t.Errorf("Expected category 'Test', got '%s'", fm.Category)
	}

	if len(fm.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(fm.Tags))
	}

	if body != "Body content\nLine 2" {
		t.Errorf("Expected body 'Body content\\nLine 2', got '%s'", body)
	}
}

func TestParseContent_CRLF(t *testing.T) {
	// Test with CRLF line endings
	content := []byte("---\r\ntitle: Test Article\r\ncategory: Test\r\ntags:\r\n  - tag1\r\n  - tag2\r\nwip: false\r\n---\r\n\r\nBody content\r\nLine 2")

	fm, body, err := ParseContent(content)
	if err != nil {
		t.Fatalf("ParseContent failed: %v", err)
	}

	if fm.Title != "Test Article" {
		t.Errorf("Expected title 'Test Article', got '%s'", fm.Title)
	}

	if fm.Category != "Test" {
		t.Errorf("Expected category 'Test', got '%s'", fm.Category)
	}

	if len(fm.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(fm.Tags))
	}

	// Body should be trimmed, but CRLF should be preserved in content
	if body == "" {
		t.Error("Body should not be empty")
	}
}

func TestParseContent_Invalid(t *testing.T) {
	// Test with invalid content (no front matter)
	content := []byte("Just some content without front matter")

	_, _, err := ParseContent(content)
	if err == nil {
		t.Error("ParseContent should fail for content without front matter")
	}

	// Check that error message mentions line endings
	if err != nil && !contains(err.Error(), "line endings") {
		t.Errorf("Error message should mention line endings, got: %v", err)
	}
}

func TestParseContent_RoundTrip(t *testing.T) {
	// Test that GenerateContent and ParseContent work together
	originalFm := types.FrontMatter{
		Title:           "Round Trip Test",
		Category:        "Test/Category",
		Tags:            []string{"test", "roundtrip"},
		Wip:             true,
		RemoteUpdatedAt: "2025-01-01T00:00:00Z",
	}
	originalBody := "This is a test body\nWith multiple lines\nAnd some content"

	// Generate content
	content, err := GenerateContent(originalFm, originalBody)
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	// Parse it back
	parsedFm, parsedBody, err := ParseContent(content)
	if err != nil {
		t.Fatalf("ParseContent failed: %v", err)
	}

	// Compare
	if parsedFm.Title != originalFm.Title {
		t.Errorf("Title mismatch: expected '%s', got '%s'", originalFm.Title, parsedFm.Title)
	}

	if parsedFm.Category != originalFm.Category {
		t.Errorf("Category mismatch: expected '%s', got '%s'", originalFm.Category, parsedFm.Category)
	}

	if len(parsedFm.Tags) != len(originalFm.Tags) {
		t.Errorf("Tags count mismatch: expected %d, got %d", len(originalFm.Tags), len(parsedFm.Tags))
	}

	if parsedFm.Wip != originalFm.Wip {
		t.Errorf("Wip mismatch: expected %v, got %v", originalFm.Wip, parsedFm.Wip)
	}

	if parsedFm.RemoteUpdatedAt != originalFm.RemoteUpdatedAt {
		t.Errorf("RemoteUpdatedAt mismatch: expected '%s', got '%s'", originalFm.RemoteUpdatedAt, parsedFm.RemoteUpdatedAt)
	}

	if parsedBody != originalBody {
		t.Errorf("Body mismatch: expected '%s', got '%s'", originalBody, parsedBody)
	}
}

// Helper functions
func startsWith(s, prefix []byte) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := range prefix {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestGenerateContent_EndsWithSingleNewline(t *testing.T) {
	tests := map[string]string{
		"改行なし":    "Body",
		"改行1つ":    "Body\n",
		"複数の改行":   "Body\n\n",
		"CRLFの改行": "Body\r\n",
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			content, err := GenerateContent(types.FrontMatter{Title: "T"}, body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasSuffix(string(content), "\n\nBody\n") {
				t.Errorf("content = %q", content)
			}
		})
	}
}

func TestParseContent_KeepsFirstLineIndent(t *testing.T) {
	_, body, err := ParseContent([]byte("---\ntitle: T\n---\n\n    code\ntext\n"))
	if err != nil {
		t.Fatal(err)
	}
	if body != "    code\ntext" {
		t.Errorf("body = %q", body)
	}
}

func TestLocalImages(t *testing.T) {
	body := "![a](./a.png)\n![a2](./a.png \"title\")\n![b](images/b.JPG)\n" +
		"![url](https://example.com/c.png)\n![data](data:image/png;base64,xx.png)\n![pdf](d.pdf)\n" +
		"`![inline](e.png)`\n```\n![fence](f.png)\n```\n![after](g.gif)"
	got := strings.Join(LocalImages(body), ",")
	if got != "./a.png,images/b.JPG,g.gif" {
		t.Errorf("images = %s", got)
	}
}

func TestReplaceLocalImagesKeepsOtherText(t *testing.T) {
	body := "text ![a](a.png \"t\") end\r\n```\n![a](a.png)\n```\n"
	got := ReplaceLocalImages(body, func(string) string { return "https://x/a.png" })
	want := "text ![a](https://x/a.png \"t\") end\r\n```\n![a](a.png)\n```\n"
	if got != want {
		t.Errorf("got %q", got)
	}
}

func TestLocalImagesInImgTag(t *testing.T) {
	body := "<img src=\"./a.png\" width=\"600\">\n<IMG width='300' src='b.jpg' />\n" +
		"<img src=\"https://example.com/c.png\">\n`<img src=\"d.png\">`\n```\n<img src=\"e.png\">\n```\n" +
		"![f](f.png) <img src=\"g.gif\" alt=\"g\">"
	got := strings.Join(LocalImages(body), ",")
	if got != "./a.png,b.jpg,f.png,g.gif" {
		t.Errorf("images = %s", got)
	}
}

func TestReplaceLocalImagesKeepsImgTagAttributes(t *testing.T) {
	body := "<img alt=\"設定画面\" src=\"./a.png\" width=\"600\">\n"
	got := ReplaceLocalImages(body, func(string) string { return "https://x/a.png" })
	want := "<img alt=\"設定画面\" src=\"https://x/a.png\" width=\"600\">\n"
	if got != want {
		t.Errorf("got %q", got)
	}
}

func TestImageFilePath(t *testing.T) {
	if got := ImageFilePath("dir/1-a.md", "img/a%20b.png"); got != "dir/img/a b.png" {
		t.Errorf("got %s", got)
	}
	if got := ImageFilePath("dir/1-a.md", "/abs/a.png"); got != "/abs/a.png" {
		t.Errorf("got %s", got)
	}
}
