# OhMyMem

**"Your Memory, Under Your Control."** **OhMyMem** is a CLI-driven, lightweight [MCP](https://modelcontextprotocol.io) server that empowers AI coding assistants with **Git-native** memory storage. We are building a universal, vendor-agnostic memory layer for the AI era—because your long-term intelligence should belong to you, not be siloed within a single platform.


It helps AI assistants remember important context across conversations by persisting constraints, decisions, patterns, and notes to your project's memory file.

---

## ✨ Features

- 🧠 **Persistent Memory** - Use MCP tools automate store project context into `.ohmymem/memory.md`
- 🔍 **Smart Project Detection** - Auto-detects projects and applies relevant templates
- 📝 **Structured Categories** - Organize memories into:
  - `constraints` - Technical requirements and rules
  - `decisions` - Architecture and design decisions
  - `patterns` - Coding patterns and conventions
  - `anti-patterns` - Things to avoid
  - `note` - General notes (fallback/default category)
- 🔄 **MCP Protocol** - Standard MCP Server use Stdio
- 📦 **Template System** - Fetch context-aware templates from GitHub/Gitee

---

## 📦 Installation

### Build from Source

```bash
git clone https://github.com/herewei/ohmymem-core.git
cd ohmymem-core
go build -o ohmymem .
```

### Or use Makefile

```bash
# macOS
make build-darwin

# Linux
make build-linux
```

### Setup Alias and PATH

Add the binary to your shell PATH for global access:

**Bash (~/.bashrc) or Zsh (~/.zshrc):**

```bash
# OhMyMem
alias ohmymem='/path/to/ohmymem-core/ohmymem'
export PATH="/path/to/ohmymem-core:$PATH"
```

Then reload your shell:

```bash
source ~/.bashrc  # or source ~/.zshrc
```

---

## 🚀 Quick Start

### 1. Initialize in Your Project

```bash
cd /path/to/your/project
ohmymem init
```

This creates:

- `.ohmymem/memory.md` - The memory storage file
- `AGENTS.md` - AI agent guidance document
- `.cursorrules` → symlink to `AGENTS.md`
- `CLAUDE.md` → symlink to `AGENTS.md`

**Options:**

```bash
ohmymem init --yes        # Skip prompts
ohmymem init --force      # Overwrite existing files
ohmymem init --repo URL   # Use custom template repository
```

### 2. Configure MCP Client

#### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "ohmymem": {
      "command": "/absolute/path/to/ohmymem",
      "args": ["mcp"],
      "cwd": "/path/to/your/project"
    }
  }
}
```

#### Other MCP Clients

Configure your client to run:

```bash
ohmymem mcp
```

---

## 🛠️ MCP Tools

Once connected, AI agents can use these tools:

### `ohmymem_read`

Read the entire memory file.

```json
{
  "name": "ohmymem_read",
  "description": "Read the working memory file (.ohmymem/memory.md)"
}
```

### `ohmymem_capture`

Add a new entry to the memory.

```json
{
  "name": "ohmymem_capture",
  "description": "Capture a new entry to the working memory",
  "parameters": {
    "category": {
      "type": "string",
      "enum": ["constraints", "decisions", "patterns", "anti-patterns", "note"],
      "description": "Category for the entry. Defaults to 'note' if not specified."
    },
    "tag": {
      "type": "string",
      "required": true,
      "description": "Tag for the entry (max 50 chars, auto-wrapped in brackets)"
    },
    "content": {
      "type": "string",
      "required": true,
      "description": "Content to remember (max 2000 chars)"
    },
    "rationale": {
      "type": "string",
      "description": "Optional reason/justification (max 500 chars)"
    }
  }
}
```

---

## 📁 Project Structure

```
your-project/
├── .ohmymem/
│   ├── memory.md       # Memory storage (auto-managed)
│   └── ohmymem.log     # Debug logs
├── AGENTS.md           # AI guidance document
├── .cursorrules        # → symlink to AGENTS.md
└── CLAUDE.md           # → symlink to AGENTS.md
```

### Memory File Format

```markdown
### Constraints

<!-- entry-id: 018e... , tag: [API], time: 2026-01-... -->
* **[API]** Use RESTful conventions for all endpoints
<!-- entry-end -->

### Decisions

<!-- entry-id: 018e... , tag: [DB], time: 2026-01-... -->
* **[DB]** Use PostgreSQL as primary database (*Rationale: ACID compliance*)
<!-- entry-end -->

### Note

<!-- entry-id: 018e... , tag: [TODO], time: 2026-01-... -->
* **[TODO]** Review authentication module next week
<!-- entry-end -->
```

---

## 🔧 Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OHMYMEM_DEBUG` | Enable debug logging (`true`/`false`) |

### Template Repositories

Default templates are fetched from:

- GitHub: `https://github.com/herewei/ohmymem-templates.git`
- Gitee: `https://gitee.com/herewei/ohmymem-templates.git`

Use custom repo:

```bash
ohmymem init --repo https://github.com/your/templates.git
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│           MCP Client                    │
│     (Claude Desktop / Cursor / ...)     │
└─────────────────┬───────────────────────┘
                  │ stdio
┌─────────────────▼───────────────────────┐
│         OhMyMem MCP Server              │
│  ┌─────────────────────────────────┐    │
│  │  Tools: read / capture          │    │
│  └─────────────────────────────────┘    │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│      Domain Layer (MemoryService)       │
│  - Validation, Template Rendering       │
│  - Entry Preparation                    │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│   Infrastructure Layer                  │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │ File Store  │  │ Git Repo Fetch  │   │
│  │ (flock)     │  │ (Templates)     │   │
│  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────┘
```

---

## 🧪 Development

```bash
# Run tests
go test ./...

# Run e2e tests
go test -v ./tests/e2e/...

# Build binary
go build -o ohmymem .

# Run directly
go run .
```

---

## 📄 License

Apache-2.0 License - see [LICENSE](LICENSE) file for details.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## 🙏 Acknowledgments

- [MCP Protocol](https://modelcontextprotocol.io) - Model Context Protocol
- [mcp-go](https://github.com/mark3labs/mcp-go) - Go SDK for MCP
- [Cobra](https://github.com/spf13/cobra) - CLI framework
