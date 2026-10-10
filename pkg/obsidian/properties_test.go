package obsidian

import (
	"reflect"
	"testing"
)

func TestGetFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    string
		wantErr bool
	}{
		{"no frontmatter", "# Title\nbody", "", false},
		{"simple", "---\ntype: journal\n---\nbody", "type: journal", false},
		{"horizontal rule in body", "---\na: 1\n---\nfoo\n---\nbar", "a: 1", false},
		{"dashes inside value", "---\ntitle: a---b\n---\n", "title: a---b", false},
		{"crlf", "---\r\na: 1\r\n---\r\n", "a: 1", false},
		{"dot terminator", "---\na: 1\n...\n", "a: 1", false},
		{"unclosed", "---\na: 1\nbody", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFrontmatter(tt.text)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseProperty(t *testing.T) {
	tests := []struct {
		in   string
		want Property
	}{
		{"done: true", Property{Name: "done", Content: "true", Value: 1, Type: Boolean}},
		{"done: false", Property{Name: "done", Content: "false", Type: Boolean}},
		{"reading: 2 hours", Property{Name: "reading", Content: "2 hours", Value: 7200, Unit: "seconds", Type: Numeric}},
		{"run: 30 minutes", Property{Name: "run", Content: "30 minutes", Value: 1800, Unit: "seconds", Type: Numeric}},
		{"pages: 12", Property{Name: "pages", Content: "12", Value: 12, Type: Numeric}},
		{"distance: 5 km", Property{Name: "distance", Content: "5 km", Value: 5, Unit: "km", Type: Numeric}},
		{"wake: 07:30", Property{Name: "wake", Content: "07:30", Type: Text}},
		{"mood: Good", Property{Name: "mood", Content: "good", Type: Text}},
		{"empty:", Property{Name: "empty"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseProperty(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGetProperties(t *testing.T) {
	text := `---
type: journal
tags:
  - daily
  - health
reading: 1 hour
wake: "07:30"
empty:
---
# Body

---

more body
`
	got, err := GetProperties(text)
	if err != nil {
		t.Fatal(err)
	}
	want := []Property{
		{Name: "type", Content: "journal", Type: Text},
		{Name: "tags", Content: "daily, health", Type: Text},
		{Name: "reading", Content: "1 hour", Value: 3600, Unit: "seconds", Type: Numeric},
		{Name: "wake", Content: "07:30", Type: Text},
		{Name: "empty"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

func TestGetPropertiesInvalidYAML(t *testing.T) {
	if _, err := GetProperties("---\na: [unclosed\n---\n"); err == nil {
		t.Error("expected error for invalid YAML")
	}
}
