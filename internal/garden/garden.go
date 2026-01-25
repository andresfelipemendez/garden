package garden

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
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
	ModTime   time.Time
}

type BacklinkData struct {
	Href  string
	Title string
}

type TemplateData struct {
	Title       string
	Content     template.HTML
	Backlinks   []BacklinkData
	Notes       []NoteLink
	CurrentHref string
}

type IndexTemplateData struct {
	Title string
	Notes []NoteLink
}

type NoteLink struct {
	Href    string
	Title   string
	Updated string
}

func Parse(markdown []byte) (*ParseResult, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			&wikilink.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					html.WithClasses(true),
				),
			),
		),
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
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/css/style.css">
    <link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@24,400,0,0&icon_names=dock_to_right,potted_plant" />
</head>
<body class="sidebar-expanded">
    <aside class="sidebar">
        <header class="sidebar-header">
            <a href="/" class="site-title"><span class="material-symbols-outlined">potted_plant</span>Note Garden</a>
            <button class="sidebar-toggle" onclick="toggleSidebar()" aria-label="Toggle sidebar">
                <span class="material-symbols-outlined">dock_to_right</span>
            </button>
        </header>
        <nav class="backlinks">
            <h2>Linked from</h2>
            {{if .Backlinks}}
            <ul>
            {{range .Backlinks}}
                <li><a href="{{.Href}}">{{.Title}}</a></li>
            {{end}}
            </ul>
            {{else}}
            <p class="empty">No backlinks yet</p>
            {{end}}
        </nav>
        <nav class="all-notes">
            <div class="notes-header">
                <h2>Notes</h2>
                <div class="sort-toggle">
                    <button onclick="sortNotes('date')" class="sort-btn active" data-sort="date">recent</button>
                    <button onclick="sortNotes('alpha')" class="sort-btn" data-sort="alpha">a-z</button>
                </div>
            </div>
            <ul id="notes-list">
            {{$current := .CurrentHref}}
            {{range .Notes}}
                <li data-title="{{.Title}}" data-date="{{.Updated}}"><a href="{{.Href}}"{{if eq .Href $current}} class="active"{{end}}>{{.Title}}</a></li>
            {{end}}
            </ul>
        </nav>
    </aside>
    <main>
        <article>
            {{.Content}}
        </article>
    </main>
    <script>
        function toggleSidebar() {
            document.body.classList.toggle('sidebar-expanded');
        }

        function sortNotes(mode) {
            var currentActive = document.querySelector('.sort-btn.active');
            if (currentActive && currentActive.dataset.sort === mode) {
                mode = mode === 'date' ? 'alpha' : 'date';
            }
            var list = document.getElementById('notes-list');
            var items = Array.from(list.querySelectorAll('li'));
            items.sort(function(a, b) {
                if (mode === 'alpha') {
                    return a.dataset.title.localeCompare(b.dataset.title);
                } else {
                    return b.dataset.date.localeCompare(a.dataset.date);
                }
            });
            items.forEach(function(item) { list.appendChild(item); });
            document.querySelectorAll('.sort-btn').forEach(function(btn) {
                btn.classList.toggle('active', btn.dataset.sort === mode);
            });
        }

        document.querySelectorAll('pre.chroma').forEach(function(pre) {
            var btn = document.createElement('button');
            btn.className = 'copy-btn';
            btn.innerHTML = '<span class="material-icons">content_copy</span>';
            btn.onclick = function() {
                var code = pre.querySelector('code');
                var text = code.innerText.trim();
                navigator.clipboard.writeText(text).then(function() {
                    btn.innerHTML = '<span class="material-icons">check</span>';
                    setTimeout(function() {
                        btn.innerHTML = '<span class="material-icons">content_copy</span>';
                    }, 2000);
                });
            };
            pre.appendChild(btn);
        });
    </script>
</body>
</html>
`))

var md = goldmark.New(
	goldmark.WithExtensions(
		&wikilink.Extender{},
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(
				html.WithClasses(true),
			),
		),
	),
)

//go:embed index.tmpl
var indexTmplContent string
var indexTmpl = template.Must(template.New("index").Parse(indexTmplContent))

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

	modTime := getGitModTime(path)

	return &Note{
		Slug:    slug,
		Title:   result.Title,
		Body:    body,
		Links:   result.Links,
		ModTime: modTime,
	}, nil
}

func getGitModTime(path string) time.Time {
	cmd := exec.Command("git", "log", "-1", "--format=%cI", "--", path)
	out, err := cmd.Output()
	if err != nil {
		return time.Now()
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out)))
	if err != nil {
		return time.Now()
	}
	return t
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

func renderNote(outDir string, note *Note, site *Site, noteLinks []NoteLink) error {
	var buf bytes.Buffer
	if err := md.Convert(note.Body, &buf); err != nil {
		return err
	}

	data := TemplateData{
		Title:       note.Title,
		Content:     template.HTML(buf.String()),
		Backlinks:   resolveBacklinks(note.Backlinks, site),
		Notes:       noteLinks,
		CurrentHref: note.Slug + ".html",
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

func renderIndex(outDir string, site *Site) error {
	var notes []*Note
	for _, note := range site.Notes {
		notes = append(notes, note)
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].ModTime.After(notes[j].ModTime)
	})

	var noteLinks []NoteLink
	for _, note := range notes {
		noteLinks = append(noteLinks, NoteLink{
			Href:    note.Slug + ".html",
			Title:   note.Title,
			Updated: note.ModTime.Format("02 Jan 2006"),
		})
	}

	data := IndexTemplateData{
		Title: "Garden",
		Notes: noteLinks,
	}

	outPath := filepath.Join(outDir, "index.html")
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return indexTmpl.Execute(f, data)
}

func Build(paths []string, outDir string) error {
	site := &Site{
		Notes:    make(map[string]*Note),
		Forward:  make(map[string][]string),
		Backward: make(map[string][]string),
	}

	for _, path := range paths {
		// Skip index.md - index is generated dynamically
		if filepath.Base(path) == "index.md" {
			continue
		}
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

	// Build sorted note links for sidebar
	var notes []*Note
	for _, note := range site.Notes {
		notes = append(notes, note)
	}
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].ModTime.After(notes[j].ModTime)
	})
	var noteLinks []NoteLink
	for _, note := range notes {
		noteLinks = append(noteLinks, NoteLink{
			Href:    note.Slug + ".html",
			Title:   note.Title,
			Updated: note.ModTime.Format("2006-01-02"),
		})
	}

	for _, note := range site.Notes {
		if err := renderNote(outDir, note, site, noteLinks); err != nil {
			return fmt.Errorf("%s: %w", note.Slug, err)
		}
	}

	if err := renderIndex(outDir, site); err != nil {
		return fmt.Errorf("index: %w", err)
	}

	return nil
}
