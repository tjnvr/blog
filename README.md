# blog

A static site generator built in Go.

## Prerequisites

* Go 1.25+
* [Task](https://taskfile.dev/)
* [browser-sync](https://browsersync.io/) (optional, for `dev` live reload)

## Setup

```bash
task setup
```

## Quick Start

```bash
# Generate the site
task generate
```

The generated site will be in the `target/build/` directory.

## Available Tasks

Run `task --list` to see all available tasks.

## Development

```bash
# Full validation before committing
task validate

# Generate and preview site locally
task dev

## Project Structure

```
blog/
├── main.go
├── styles/                      # Tailwind CSS input and optional styling config
├── scripts/                     # JavaScript files
├── internal/
│   └── generator/
│       ├── page/
│       │   ├── html/            # HTML substitution and validation
│       │   │   ├── substitution/ # {{title}}, {{content}}, {{navigation}}
│       │   │   └── validation/  # link, image, navigation, script validators
│       │   └── markdown/        # Markdown substitution (runs before HTML conversion)
│       │       └── substitution/ # {{list-child-articles}}
│       └── site/                # Site-level generation (routing, asset copying)
├── content/
│   ├── markdown/                # Markdown source files
│   └── assets/                  # Static assets (images, etc.)
└── target/build/                # Generated output
```

## Architecture

### Generation Pipeline

```
Markdown Files (content/markdown/)
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  Site Generator                                         │
│  ├─ List sections (top-level dirs in content/markdown/) │
│  │   └─ Section display name from # title in index.md  │
│  ├─ Copy assets & scripts                               │
│  └─ For each .md file → Page Generator                  │
│      ├─ Markdown Substitution Registry                  │
│      │   └─ {{list-child-articles}} → Markdown links    │
│      │       to sibling .md files (title from H1)       │
│      ├─ Markdown Converter (Goldmark + GFM)             │
│      │   └─ Style Transformer (optional)                │
│      ├─ HTML Substitution Registry                      │
│      │   ├─ {{title}} → First H1 from markdown          │
│      │   ├─ {{content}} → Converted HTML                │
│      │   └─ {{navigation}} → Nav bar with section links │
│      │       (display names from each section's index)  │
│      └─ Validators                                      │
│          ├─ Link Validator                              │
│          ├─ Image Validator                             │
│          └─ Navigation Validator                        │
└─────────────────────────────────────────────────────────┘
    │
    ▼
HTML Files (target/build/)
```

### Template Substitutions

The generator supports two layers of substitutions applied in order: markdown substitutions (before HTML conversion) and HTML substitutions (applied to the template).

### Styling System

The generator uses [Tailwind CSS v4](https://tailwindcss.com/) with the [Typography plugin](https://tailwindcss.com/docs/typography-plugin) for automatic prose styling. 
CSS is built using the Tailwind Standalone CLI. Configuration is done in `styles/input.css` using [CSS-based configuration](https://tailwindcss.com/docs/installation/tailwind-cli).

* **Default behavior**: All Markdown content is wrapped in `<article class="prose prose-lg">`, which applies consistent typography styles to headings, paragraphs, links, code blocks, etc.

* **Custom styling**: Create a `styles/styles.json` file to add CSS classes to specific elements:

```json
{
  "elements": {
    "heading1": "text-4xl font-bold",
    "image": "rounded-lg shadow-md",
    "link": "text-blue-600 hover:underline",
    "blockquote": "border-l-4 border-gray-300 italic"
  },
  "contexts": {
    "post": {
      "heading1": "text-blue-900"
    }
  }
}
```

  * Supported element keys: `heading1`, `heading2`, `heading3`, `heading4`, `heading5`, `heading6`, `paragraph`, `link`, `image`, `codeblock`, `code`, `blockquote`, `list`, `listitem`.

  * Contexts: Files in `posts/` automatically get the `post` context, allowing context-specific styling.

  * Validation: Invalid keys cause the generator to exit with an error listing valid options.

* **Inline attributes**: For precise control on specific elements, use the inline attribute syntax directly in Markdown (take precedence over `styles.json`).
