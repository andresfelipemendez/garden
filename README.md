# Garden

A static site generator for interconnected notes, written in Go.

Parses Obsidian-flavored markdown (wiki links) and produces a static HTML site with bidirectional links.

**Live site**: [garden.andresfelipemendez.com](https://garden.andresfelipemendez.com)

## Features

- **Wiki links**: `[[note]]` syntax links between notes
- **Backlinks**: Automatically generated "Linked from" section
- **Syntax highlighting**: Chroma with GitHub style, line numbers, copy button
- **Sidebar**: Toggleable navigation with backlinks
- **Fonts**: iA Writer Duo (body), Source Code Pro (code)

## Build

```bash
# Build CLI
go build -o garden ./cmd/garden

# Generate site
./garden

# Run tests
go test ./...
```

## Local Development

```bash
# Build and run with Docker
docker build -t garden:local .
docker run -p 8080:80 garden:local

# View at http://localhost:8080
```

## Structure

```
garden/
├── cmd/garden/          # CLI entry point
├── internal/garden/     # Core logic (parse, build, render)
├── notes/               # Markdown source files
├── static/
│   ├── css/style.css    # Styles + syntax highlighting
│   └── fonts/           # iA Writer Duo, Source Code Pro
├── public/              # Generated output (not committed)
├── Dockerfile
└── README.md
```

## Adding Notes

Create a markdown file in `notes/`:

```markdown
# Note Title

Content with [[wiki-links]] to other notes.

Code blocks get syntax highlighting:

```yaml
apiVersion: v1
kind: Example
```
```

The filename becomes the slug (e.g., `my-note.md` → `my-note.html`).

## Deployment

- Push to `main` triggers GitHub Actions
- Builds Docker image → `ghcr.io/andresfelipemendez/garden:latest`
- ArgoCD syncs to Kubernetes cluster
- Served via nginx at garden.andresfelipemendez.com
