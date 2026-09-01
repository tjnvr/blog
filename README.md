# blog

A static site generator built in Go.

## Prerequisites

* Go 1.25+
* [Task](https://taskfile.dev/)
* [browser-sync](https://browsersync.io/) (optional, for `dev` live reload)

## Develop the generator

```bash
# Install needed tools
task setup

# Validate the Go generator
task go-validate

# Generate the site
task generate

# Generate the site and serve it locally
task serve &
task dev

# Validate the site served locally
BASE_URL=http://localhost:3000/ task validate
```

## Write content

### Markdown syntax

[goldmark](https://github.com/yuin/goldmark) + [GFM](https://github.github.com/gfm/).

### Markdown content features

| Feature | Where | Behaviour | Example |
|---|---|---|---|
| `{{summary}}` | any page | Table of contents from the page's headings. | `{{summary}}` |
| `{{list-child-pages}}` | `index.md` page | Vertical listing of sibling `.md` pages:  date, title, summary, optional thumbnail. | `{{list-child-pages}}` |
| `{.small-centered}` | any page | Displays the image small and centered. | `![alt](img.png){.small-centered}` |

### Page metadata

A page can declare metadata via HTML comments, preferably at the top of its markdown
source, before the content.

| Comment | Meaning |
|---|---|
| `<!-- creation-date: YYYY-MM-DD -->` | Publication date. Required for `{{list-child-pages}}` |
| `<!-- description: ... -->` | Meta description tag; also the summary text in `{{list-child-pages}}`. |
| `<!-- image: relative/path.png -->` | Listing thumbnail for `{{list-child-pages}}`. Optional. |
| `<!-- seq: N -->` | Nav order. Pages without one sort last. |
