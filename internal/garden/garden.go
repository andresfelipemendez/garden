package garden

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/wikilink"
)

type ParseResult struct {
	Title string
	HTML  []byte
	Links []string
}

type Site struct {
	Notes    map[string]*Note
	Forward  map[string][]string
	Backward map[string][]string
}

type Note struct {
	Slug      string
	Title     string
	Body      []byte
	Links     []string
	Backlinks []string
}

type BacklinkData struct {
	Href  string
	Title string
}

type TemplateData struct {
	Title     string
	Content   template.HTML
	Backlinks []BacklinkData
}

func Parse(markdown []byte) (*ParseResult, error) {
	md := goldmark.New(
		goldmark.WithExtensions(&wikilink.Extender{}),
	)

	reader := text.NewReader(markdown)
	doc := md.Parser().Parse(reader)

	var title string
	var links []string
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if wl, ok := n.(*wikilink.Node); ok {
			links = append(links, string(wl.Target))
			return ast.WalkContinue, nil
		}

		if title == "" {
			if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
				var buf bytes.Buffer
				for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
					if txt, ok := child.(*ast.Text); ok {
						buf.Write(txt.Value(markdown))
					}
				}
				title = buf.String()
				return ast.WalkSkipChildren, nil
			}
		}

		return ast.WalkContinue, nil
	})

	var html bytes.Buffer
	if err := md.Renderer().Render(&html, markdown, doc); err != nil {
		return nil, err
	}

	return &ParseResult{
		Title: title,
		HTML:  html.Bytes(),
		Links: links,
	}, nil
}

var tmpl = template.Must(template.New("note").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/css/style.css">
</head>
<body>
    <article>
        {{.Content}}
    </article>
    {{if .Backlinks}}
    <nav class="backlinks">
        <h2>Linked from</h2>
        <ul>
        {{range .Backlinks}}
            <li><a href="{{.Href}}">{{.Title}}</a></li>
        {{end}}
        </ul>
    </nav>
    {{end}}
</body>
</html>
`))

var md = goldmark.New(
	goldmark.WithExtensions(&wikilink.Extender{}),
)

func parseNote(path string) (*Note, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result, err := Parse(body)
	if err != nil {
		return nil, err
	}

	slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	return &Note{
		Slug:  slug,
		Title: result.Title,
		Body:  body,
		Links: result.Links,
	}, nil
}

func buildBacklinks(site *Site) {
	for src, dests := range site.Forward {
		for _, dest := range dests {
			site.Backward[dest] = append(site.Backward[dest], src)
		}
	}
}

func resolveBacklinks(slugs []string, site *Site) []BacklinkData {
	var result []BacklinkData
	for _, slug := range slugs {
		note, ok := site.Notes[slug]
		if !ok {
			continue
		}
		result = append(result, BacklinkData{
			Href:  slug + ".html",
			Title: note.Title,
		})
	}
	return result
}

func renderNote(outDir string, note *Note, site *Site) error {
	var buf bytes.Buffer
	if err := md.Convert(note.Body, &buf); err != nil {
		return err
	}

	data := TemplateData{
		Title:     note.Title,
		Content:   template.HTML(buf.String()),
		Backlinks: resolveBacklinks(note.Backlinks, site),
	}

	outPath := filepath.Join(outDir, note.Slug+".html")
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func Build(paths []string, outDir string) error {
	site := &Site{
		Notes:    make(map[string]*Note),
		Forward:  make(map[string][]string),
		Backward: make(map[string][]string),
	}

	for _, path := range paths {
		note, err := parseNote(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		site.Notes[note.Slug] = note
		site.Forward[note.Slug] = note.Links
	}

	buildBacklinks(site)

	for _, note := range site.Notes {
		note.Backlinks = site.Backward[note.Slug]
	}

	for _, note := range site.Notes {
		if err := renderNote(outDir, note, site); err != nil {
			return fmt.Errorf("%s: %w", note.Slug, err)
		}
	}

	return nil
}
