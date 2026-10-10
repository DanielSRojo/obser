package obsidian

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestTidyFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		removed []string
	}{
		{
			name:    "template leftovers removed",
			in:      "---\ntype: journal\nguitar: 30 minutes\nread:\nmeditate: \naliases: []\n---\nbody\n",
			want:    "---\ntype: journal\nguitar: 30 minutes\n---\nbody\n",
			removed: []string{"read", "meditate", "aliases"},
		},
		{
			name:    "block list kept",
			in:      "---\ntype: journal\ntags:\n  - travel\nread:\n---\n",
			want:    "---\ntype: journal\ntags:\n  - travel\n---\n",
			removed: []string{"read"},
		},
		{
			name: "falsy values kept",
			in:   "---\nchain: false\nsexTry: 0\n---\n",
			want: "---\nchain: false\nsexTry: 0\n---\n",
		},
		{
			name:    "null forms removed",
			in:      "---\na: \"\"\nb: ~\nc: null\nd: ''\ne: {}\nf: \"~\"\n---\n",
			want:    "---\nf: \"~\"\n---\n",
			removed: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:    "empty type kept",
			in:      "---\ntype:\nread:\n---\n",
			want:    "---\ntype:\n---\n",
			removed: []string{"read"},
		},
		{
			name:    "CRLF preserved",
			in:      "---\r\ntype: journal\r\nread:\r\nguitar: 1 hour\r\n---\r\nbody\r\n",
			want:    "---\r\ntype: journal\r\nguitar: 1 hour\r\n---\r\nbody\r\n",
			removed: []string{"read"},
		},
		{
			name:    "missing final newline preserved",
			in:      "---\nread:\nguitar: 1 hour\n---\nbody",
			want:    "---\nguitar: 1 hour\n---\nbody",
			removed: []string{"read"},
		},
		{
			name:    "body untouched",
			in:      "---\nread:\n---\n```yaml\nkey:\nread:\n---\n```\n",
			want:    "---\n---\n```yaml\nkey:\nread:\n---\n```\n",
			removed: []string{"read"},
		},
		{
			name:    "no properties left keeps delimiters",
			in:      "---\naliases: []\ntags: []\n---\n",
			want:    "---\n---\n",
			removed: []string{"aliases", "tags"},
		},
		{
			name:    "comments",
			in:      "---\nread:\n  # indented goes\n# unindented stays\n\nguitar: 1 hour\nwrite:\n# last stays\n---\n",
			want:    "---\n# unindented stays\n\nguitar: 1 hour\n# last stays\n---\n",
			removed: []string{"read", "write"},
		},
		{
			name: "no frontmatter",
			in:   "read:\n",
			want: "read:\n",
		},
		{
			name: "empty frontmatter",
			in:   "---\n---\n",
			want: "---\n---\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, removed, err := TidyFrontmatter([]byte(tt.in))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
			if !reflect.DeepEqual(removed, tt.removed) {
				t.Errorf("removed %v, want %v", removed, tt.removed)
			}

			again, removed, err := TidyFrontmatter(got)
			if err != nil {
				t.Fatal(err)
			}
			if string(again) != string(got) || removed != nil {
				t.Errorf("second pass changed the content: %q, removed %v", again, removed)
			}
		})
	}
}

func TestTidyFrontmatterErrors(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "unterminated", in: "---\nread:\nbody\n", wantErr: "closing frontmatter delimiter"},
		{name: "invalid YAML", in: "---\nread: [\n---\n", wantErr: "invalid frontmatter"},
		{name: "duplicate key", in: "---\nread:\nread:\n---\n", wantErr: "invalid frontmatter"},
		{name: "templater", in: "---\ndate: <% tp.date.now() %>\nread:\n---\n", wantErr: "Templater"},
		{name: "flow mapping", in: "---\n{read: , guitar: 1}\n---\n", wantErr: "invalid frontmatter"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := TidyFrontmatter([]byte(tt.in))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want an error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestTidyJournal(t *testing.T) {
	entry := "---\ntype: journal\nguitar: 30 minutes\nread:\n---\nbody\n"
	setupVault(t, map[string]string{
		"2025-03-13.md":        entry,
		"2025-03-14.md":        entry,
		"2025-03-15.md":        entry, // the cutoff day ("today")
		"2025-03-16.md":        entry, // future, created by fill
		"2025-03-12.md":        "---\nguitar: 30 minutes\nread:\n---\n",
		"2023-11-29-guitar.md": "---\ntype: journal\ntags:\n---\n",
		"2025-01-02.md":        "---\ntype: journal\nread:\n",
		"2025-01-03.md":        "---\ntype: journal\ndate: <% tp.date.now() %>\nread:\n---\n",
		"2025-01-04.md":        "---\ntype: journal\nguitar: 1 hour\n---\n",
	})
	path := func(name string) string { return filepath.Join(vaultDir, name) }

	mtime := time.Date(2025, 3, 15, 21, 30, 0, 0, time.Local)
	for _, name := range []string{"2025-03-13.md", "2025-03-14.md", "2025-03-15.md", "2025-03-16.md"} {
		if err := os.Chtimes(path(name), mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}
	before := time.Date(2025, 3, 15, 12, 0, 0, 0, time.Local)

	results, _, err := TidyJournal(before, true)
	if err != nil {
		t.Fatal(err)
	}
	wantResults := []TidyResult{
		{Path: path("2025-03-13.md"), Removed: []string{"read"}},
		{Path: path("2025-03-14.md"), Removed: []string{"read"}},
	}
	if !reflect.DeepEqual(results, wantResults) {
		t.Errorf("dry run results %v, want %v", results, wantResults)
	}
	assertEntry(t, path("2025-03-14.md"), entry, mtime)

	ino := inode(t, path("2025-03-14.md"))
	results, skipped, err := TidyJournal(before, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(results, wantResults) {
		t.Errorf("results %v, want %v", results, wantResults)
	}
	var skippedPaths []string
	for _, s := range skipped {
		skippedPaths = append(skippedPaths, s.Path)
	}
	if want := []string{path("2025-01-02.md"), path("2025-01-03.md")}; !reflect.DeepEqual(skippedPaths, want) {
		t.Errorf("skipped %v, want %v", skipped, want)
	}

	tidied := "---\ntype: journal\nguitar: 30 minutes\n---\nbody\n"
	assertEntry(t, path("2025-03-13.md"), tidied, mtime)
	assertEntry(t, path("2025-03-14.md"), tidied, mtime)
	assertEntry(t, path("2025-03-15.md"), entry, mtime)
	assertEntry(t, path("2025-03-16.md"), entry, mtime)
	if got := inode(t, path("2025-03-14.md")); got != ino {
		t.Errorf("inode changed from %d to %d", ino, got)
	}
	for name, content := range map[string]string{
		"2025-03-12.md":        "---\nguitar: 30 minutes\nread:\n---\n",
		"2023-11-29-guitar.md": "---\ntype: journal\ntags:\n---\n",
	} {
		if got, err := os.ReadFile(path(name)); err != nil || string(got) != content {
			t.Errorf("%s was modified: %q, %v", name, got, err)
		}
	}

	results, _, err = TidyJournal(before, false)
	if err != nil || len(results) != 0 {
		t.Errorf("second run: %v, %v", results, err)
	}
}

func assertEntry(t *testing.T, path, content string, mtime time.Time) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("%s: got %q, want %q", filepath.Base(path), got, content)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(mtime) {
		t.Errorf("%s: mtime %v, want %v", filepath.Base(path), info.ModTime(), mtime)
	}
}

func inode(t *testing.T, path string) uint64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return uint64(info.Sys().(*syscall.Stat_t).Ino)
}
