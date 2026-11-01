package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shellme/esa-cli/internal/api"
)

var baseTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// fakeEsa 記事の状態を持つesa APIの偽物
type fakeEsa struct {
	posts     map[int]map[string]interface{}
	fetchCode int
	patches   []map[string]interface{}
	queries   []string
	perPage   int
	uploads   int
}

func newFakeEsa() *fakeEsa {
	return &fakeEsa{posts: map[int]map[string]interface{}{}, perPage: 100}
}

func (f *fakeEsa) add(number int, category string, wip bool, tags ...string) {
	f.posts[number] = map[string]interface{}{
		"number": number, "name": fmt.Sprintf("p%d", number), "full_name": fmt.Sprintf("%s/p%d", category, number),
		"wip": wip, "body_md": "body", "category": category, "tags": tags,
		"created_at": baseTime, "updated_at": baseTime,
	}
}

func (f *fakeEsa) respond(code int, v interface{}) *http.Response {
	data, _ := json.Marshal(v)
	return &http.Response{StatusCode: code, Status: strconv.Itoa(code), Header: http.Header{}, Body: io.NopCloser(bytes.NewReader(data))}
}

func (f *fakeEsa) Do(req *http.Request) (*http.Response, error) {
	if strings.HasSuffix(req.URL.Path, "/attachments") {
		f.uploads++
		url := fmt.Sprintf("https://files.esa.io/uploads/%d.png", f.uploads)
		return f.respond(201, map[string]interface{}{"attachment": map[string]interface{}{"url": url}}), nil
	}
	if match := regexp.MustCompile(`/posts/(\d+)$`).FindStringSubmatch(req.URL.Path); match != nil {
		number, _ := strconv.Atoi(match[1])
		post, ok := f.posts[number]
		if req.Method == http.MethodGet {
			if f.fetchCode != 0 {
				return f.respond(f.fetchCode, nil), nil
			}
			if !ok {
				return f.respond(404, nil), nil
			}
			return f.respond(200, post), nil
		}

		var body struct {
			Post map[string]interface{} `json:"post"`
		}
		json.NewDecoder(req.Body).Decode(&body)
		f.patches = append(f.patches, body.Post)
		for key, value := range body.Post {
			post[key] = value
		}
		return f.respond(200, post), nil
	}

	f.queries = append(f.queries, req.URL.Query().Get("q"))
	page, _ := strconv.Atoi(req.URL.Query().Get("page"))
	var all []map[string]interface{}
	for number := 1; number <= len(f.posts)*10; number++ {
		if post, ok := f.posts[number]; ok {
			all = append(all, post)
		}
	}
	start := (page - 1) * f.perPage
	end := start + f.perPage
	var next interface{}
	if end < len(all) {
		next = page + 1
	} else {
		end = len(all)
	}
	return f.respond(200, map[string]interface{}{"posts": all[start:end], "next_page": next, "total_count": len(all)}), nil
}

func chdirTemp(t *testing.T) {
	t.Helper()
	wd, _ := os.Getwd()
	os.Chdir(t.TempDir())
	t.Cleanup(func() { os.Chdir(wd) })
}

func writeArticle(t *testing.T, path, tags, updatedAt string) {
	t.Helper()
	content := fmt.Sprintf("---\ntitle: p1\ncategory: src\ntags: %s\nwip: true\nremote_updated_at: %q\n---\n\n", tags, updatedAt)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func appendBody(t *testing.T, path, body string) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	file.WriteString(body)
}

func never(string) bool { return false }

func TestBulkUpdateCategoryKeepsWip(t *testing.T) {
	fake := newFakeEsa()
	fake.add(1, "src", true, "a")
	client := api.NewClient("t", "token", fake)

	if _, err := client.BulkUpdateCategory(context.Background(), []int{1}, "dst", "msg"); err != nil {
		t.Fatal(err)
	}
	if _, sent := fake.patches[0]["wip"]; sent {
		t.Error("moveでwipを送ってはいけない")
	}
	if fake.posts[1]["wip"] != true {
		t.Error("WIP状態が変わった")
	}
}

func TestFindMovePostsIgnoresOtherCategories(t *testing.T) {
	fake := newFakeEsa()
	fake.add(1, "src/sub", true)
	fake.add(2, "srcother", false)
	client := api.NewClient("t", "token", fake)

	posts, _, err := findMovePosts(context.Background(), client, api.ListPostsOptions{Category: "src"})
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].Number != 1 {
		t.Errorf("対象外のカテゴリが含まれている: %+v", posts)
	}
	if fake.queries[0] != "in:src" {
		t.Errorf("q = %q", fake.queries[0])
	}
}

func TestUpdateFile(t *testing.T) {
	t.Run("最後のタグの削除と空の本文を送信できる", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true, "a")
		client := api.NewClient("t", "token", fake)
		writeArticle(t, "1-p1.md", "[a]", baseTime.Format(time.RFC3339))

		if err := updateFile(context.Background(), client, "1-p1.md", updateOptions{removeTags: "a"}, never); err != nil {
			t.Fatal(err)
		}
		if tags, ok := fake.patches[0]["tags"].([]interface{}); !ok || len(tags) != 0 {
			t.Errorf("tags = %v", fake.patches[0]["tags"])
		}
		if fake.patches[0]["body_md"] != "" {
			t.Errorf("body_md = %v", fake.patches[0]["body_md"])
		}
	})

	t.Run("リモート確認に失敗したら更新しない", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true)
		fake.fetchCode = 500
		client := api.NewClient("t", "token", fake)
		writeArticle(t, "1-p1.md", "[]", baseTime.Format(time.RFC3339))

		if err := updateFile(context.Background(), client, "1-p1.md", updateOptions{}, never); err == nil {
			t.Fatal("エラーになるべき")
		}
		if len(fake.patches) != 0 {
			t.Error("更新してはいけない")
		}
	})

	t.Run("競合を拒否したら更新しない", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true)
		fake.posts[1]["updated_at"] = baseTime.Add(time.Hour)
		client := api.NewClient("t", "token", fake)
		writeArticle(t, "1-p1.md", "[]", baseTime.Format(time.RFC3339))

		failed, canceled := updateAllFiles(context.Background(), client, []string{"1-p1.md"}, updateOptions{}, never)
		if failed != 0 || len(canceled) != 1 || len(fake.patches) != 0 {
			t.Errorf("failed=%d canceled=%v patches=%d", failed, canceled, len(fake.patches))
		}
	})

	t.Run("remote_updated_atがなく確認を拒否したら更新しない", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true)
		client := api.NewClient("t", "token", fake)
		os.WriteFile("1-p1.md", []byte("---\ntitle: p1\ncategory: src\nwip: true\n---\n\n"), 0644)

		err := updateFile(context.Background(), client, "1-p1.md", updateOptions{}, never)
		if !errors.Is(err, errUpdateCanceled) {
			t.Fatalf("err = %v", err)
		}
		if len(fake.patches) != 0 {
			t.Error("更新してはいけない")
		}
	})

	t.Run("remote_updated_atがなくても確認すれば更新できる", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true)
		client := api.NewClient("t", "token", fake)
		os.WriteFile("1-p1.md", []byte("---\ntitle: p1\ncategory: src\nwip: true\n---\n\n"), 0644)

		yes := func(string) bool { return true }
		if err := updateFile(context.Background(), client, "1-p1.md", updateOptions{}, yes); err != nil {
			t.Fatal(err)
		}
		if len(fake.patches) != 1 {
			t.Errorf("patches=%d", len(fake.patches))
		}
	})

	t.Run("画像をアップロードして本文をURLに置換する", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true)
		client := api.NewClient("t", "token", fake)
		os.WriteFile("a.png", []byte("png"), 0644)
		writeArticle(t, "1-p1.md", "[]", baseTime.Format(time.RFC3339))
		appendBody(t, "1-p1.md", "![a](a.png)\n![a](a.png)\n")

		if err := updateFile(context.Background(), client, "1-p1.md", updateOptions{}, never); err != nil {
			t.Fatal(err)
		}
		want := "![a](https://files.esa.io/uploads/1.png)\n![a](https://files.esa.io/uploads/1.png)"
		if fake.uploads != 1 || fake.patches[0]["body_md"] != want {
			t.Errorf("uploads=%d body=%v", fake.uploads, fake.patches[0]["body_md"])
		}
	})

	t.Run("画像の失敗で更新せず成功分はファイルに残す", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true)
		client := api.NewClient("t", "token", fake)
		os.WriteFile("a.png", []byte("png"), 0644)
		writeArticle(t, "1-p1.md", "[]", baseTime.Format(time.RFC3339))
		appendBody(t, "1-p1.md", "![a](a.png)\n![b](missing.png)\n")

		if err := updateFile(context.Background(), client, "1-p1.md", updateOptions{}, never); err == nil {
			t.Fatal("エラーになるべき")
		}
		if len(fake.patches) != 0 {
			t.Error("更新してはいけない")
		}
		content, _ := os.ReadFile("1-p1.md")
		if !strings.Contains(string(content), "uploads/1.png") || !strings.Contains(string(content), "remote_updated_at") {
			t.Errorf("content = %s", content)
		}
	})

	t.Run("競合を拒否したら画像をアップロードしない", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true)
		fake.posts[1]["updated_at"] = baseTime.Add(time.Hour)
		client := api.NewClient("t", "token", fake)
		os.WriteFile("a.png", []byte("png"), 0644)
		writeArticle(t, "1-p1.md", "[]", baseTime.Format(time.RFC3339))
		appendBody(t, "1-p1.md", "![a](a.png)\n")

		updateFile(context.Background(), client, "1-p1.md", updateOptions{}, never)
		if fake.uploads != 0 {
			t.Errorf("uploads=%d", fake.uploads)
		}
	})

	t.Run("サブフォルダのファイルを更新できる", func(t *testing.T) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true)
		client := api.NewClient("t", "token", fake)
		os.Mkdir("sub", 0755)
		path := filepath.Join("sub", "1-p1.md")
		writeArticle(t, path, "[]", baseTime.Format(time.RFC3339))

		if err := updateFile(context.Background(), client, path, updateOptions{}, never); err != nil {
			t.Fatal(err)
		}
	})
}

func TestFindArticleFiles(t *testing.T) {
	chdirTemp(t)
	os.Mkdir("sub", 0755)
	for _, path := range []string{"1-a.md", "2-b.md", "memo.md", "sub/3-c.md"} {
		os.WriteFile(path, nil, 0644)
	}

	files, err := findArticleFiles([]string{"*.md", "1-*.md"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(files, ",") != "1-a.md,2-b.md" {
		t.Errorf("files = %v", files)
	}
}

func TestFetchAllPosts(t *testing.T) {
	setup := func(t *testing.T) (*fakeEsa, *api.Client) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.perPage = 2
		for number := 1; number <= 5; number++ {
			fake.add(number, "src", false)
		}
		return fake, api.NewClient("t", "token", fake)
	}
	count := func() int {
		files, _ := filepath.Glob("*.md")
		return len(files)
	}

	t.Run("全ページを取得する", func(t *testing.T) {
		_, client := setup(t)
		if _, err := fetchAllPosts(context.Background(), client, fetchAllOptions{}); err != nil {
			t.Fatal(err)
		}
		if count() != 5 {
			t.Errorf("count = %d", count())
		}
	})

	t.Run("limitを守る", func(t *testing.T) {
		_, client := setup(t)
		fetchAllPosts(context.Background(), client, fetchAllOptions{limit: 3, category: "src"})
		if count() != 3 {
			t.Errorf("count = %d", count())
		}
	})

	t.Run("dry-runでは保存しない", func(t *testing.T) {
		_, client := setup(t)
		fetchAllPosts(context.Background(), client, fetchAllOptions{dryRun: true})
		if count() != 0 {
			t.Errorf("count = %d", count())
		}
	})

	t.Run("既存ファイルは上書きしない", func(t *testing.T) {
		_, client := setup(t)
		os.WriteFile("1-p1.md", []byte("local"), 0644)
		fetchAllPosts(context.Background(), client, fetchAllOptions{})
		data, _ := os.ReadFile("1-p1.md")
		if string(data) != "local" {
			t.Error("上書きされた")
		}
		fetchAllPosts(context.Background(), client, fetchAllOptions{overwrite: true})
		data, _ = os.ReadFile("1-p1.md")
		if string(data) == "local" {
			t.Error("--overwriteで上書きされなかった")
		}
	})

	t.Run("タイトルのスラッシュを置換する", func(t *testing.T) {
		fake, client := setup(t)
		fake.posts[1]["name"] = "a/b"
		if failed, _ := fetchAllPosts(context.Background(), client, fetchAllOptions{}); failed != 0 {
			t.Errorf("failed = %d", failed)
		}
		if _, err := os.Stat("1-a-b.md"); err != nil {
			t.Error(err)
		}
	})
}

func TestCompareArticle(t *testing.T) {
	setup := func(t *testing.T, body string) (*fakeEsa, *api.Client) {
		chdirTemp(t)
		fake := newFakeEsa()
		fake.add(1, "src", true, "a", "b")
		fake.posts[1]["body_md"] = "line1\r\nline2\r\n"
		writeArticle(t, "1-p1.md", "[b, a]", baseTime.Format(time.RFC3339))
		appendBody(t, "1-p1.md", body)
		return fake, api.NewClient("t", "token", fake)
	}
	check := func(t *testing.T, client *api.Client, want string) {
		t.Helper()
		state, _, _, err := compareArticle(context.Background(), client, "1-p1.md")
		if err != nil {
			t.Fatal(err)
		}
		if state != want {
			t.Errorf("state = %s", state)
		}
	}

	t.Run("改行コードとタグ順の違いは変更なし", func(t *testing.T) {
		_, client := setup(t, "line1\nline2\n")
		check(t, client, stateClean)
	})
	t.Run("ローカルのみ変更", func(t *testing.T) {
		_, client := setup(t, "edited\n")
		check(t, client, stateLocalChanged)
	})
	t.Run("esa側が新しく内容が違えば要確認", func(t *testing.T) {
		fake, client := setup(t, "edited\n")
		fake.posts[1]["updated_at"] = baseTime.Add(time.Hour)
		check(t, client, stateRemoteNewer)
	})
}

func TestBackupFile(t *testing.T) {
	chdirTemp(t)
	os.WriteFile("1-p1.md", []byte("local"), 0644)
	if err := backupFile("1-p1.md"); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(backupDir, "1-p1.md.*"))
	if len(files) != 1 {
		t.Fatalf("files = %v", files)
	}
	if content, _ := os.ReadFile(files[0]); string(content) != "local" {
		t.Errorf("content = %s", content)
	}
}
