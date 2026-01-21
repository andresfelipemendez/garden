package garden

import (
	"bytes"
	"testing"
)

func TestParseHeading(t *testing.T) {
	input := []byte("# Hello")

	result, err := Parse(input)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Title != "Hello" {
		t.Errorf("Title = %q, want %q", result.Title, "Hello")
	}

	if !bytes.Contains(result.HTML, []byte("<h1>Hello</h1>\n")) {
		t.Errorf("HTML = %q, want h1 tag", result.HTML)
	}
}

func TestParseExtractsWikiLinks(t *testing.T) {
	input := []byte("See [[other-note]] for details.")

	result, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Links) != 1 || result.Links[0] != "other-note" {
		t.Errorf("Links = %v, want [other-note]", result.Links)
	}
}

func TestParseTitleAndMultipleLinks(t *testing.T) {
	input := []byte(`# My Note

This references [[first-link]] and also [[second-link]].
`)

	result, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Title != "My Note" {
		t.Errorf("Title = %q, want %q", result.Title, "My Note")
	}

	if len(result.Links) != 2 {
		t.Fatalf("Links count = %d, want 2", len(result.Links))
	}
	if result.Links[0] != "first-link" {
		t.Errorf("Links[0] = %q, want %q", result.Links[0], "first-link")
	}
	if result.Links[1] != "second-link" {
		t.Errorf("Links[1] = %q, want %q", result.Links[1], "second-link")
	}

	if !bytes.Contains(result.HTML, []byte("<h1>My Note</h1>")) {
		t.Errorf("HTML missing h1, got: %s", result.HTML)
	}
	if !bytes.Contains(result.HTML, []byte(`<a href="first-link.html">first-link</a>`)) {
		t.Errorf("HTML missing first-link anchor, got: %s", result.HTML)
	}
	if !bytes.Contains(result.HTML, []byte(`<a href="second-link.html">second-link</a>`)) {
		t.Errorf("HTML missing second-link anchor, got: %s", result.HTML)
	}
}

func TestBuildBacklinks(t *testing.T) {
	site := &Site{
		Notes: map[string]*Note{
			"a": {Slug: "a", Title: "Note A"},
			"b": {Slug: "b", Title: "Note B"},
			"c": {Slug: "c", Title: "Note C"},
		},
		Forward: map[string][]string{
			"a": {"b", "c"},
			"b": {"c"},
		},
		Backward: make(map[string][]string),
	}

	buildBacklinks(site)

	if len(site.Backward["b"]) != 1 || site.Backward["b"][0] != "a" {
		t.Errorf("Backward[b] = %v, want [a]", site.Backward["b"])
	}

	if len(site.Backward["c"]) != 2 {
		t.Errorf("Backward[c] = %v, want [a, b]", site.Backward["c"])
	}
}
