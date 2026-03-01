# 🎧 DupClean — Duplicate File Cleaner

A fast, content-aware duplicate file scanner for music producers and DJs.

**The key insight:** Two files can have completely different names but identical audio content.  
DupClean hashes file *contents* — not names — so it catches every duplicate no matter how it was renamed.

---

## Installation

### macOS

**Option 1: DMG (Recommended)**
1. Download `dupclean.dmg` from releases
2. Double-click to mount
3. Drag `DupClean.app` to Applications
4. Run from Terminal: `dupclean` (adds to PATH automatically on first run)

**Option 2: Homebrew**
```bash
# Coming soon - submit a PR if you'd like to add it!
```

**Option 3: Binary**
```bash
# Download and extract
tar -xzf dupclean-darwin-arm64.tar.gz
./dupclean

# Or copy to PATH
sudo mv dupclean /usr/local/bin/
```

---

### Linux

**Binary**
```bash
# Download and extract
tar -xzf dupclean-linux-arm64.tar.gz

# Run directly
./dupclean

# Or copy to PATH
sudo mv dupclean /usr/local/bin/
```

**Required dependencies:**
- GTK (usually pre-installed on most distros)
- On headless servers: X11 forwarding or a desktop environment

---

### Windows

**Binary**
```bash
# Download and extract the zip
# Run directly
dupclean.exe
```

Or double-click in File Explorer.

---

## Usage

### CLI Mode

```bash
# Scan a folder for duplicate audio files
dupclean ~/Music/Samples

# Also include non-audio files
dupclean ~/Music/Samples --all

# Show help
dupclean --help
```

### GUI Mode (macOS/Windows/Linux with display)

```bash
# Launch GUI
dupclean
# or
dupclean --gui
```

---

## How it works

1. **Walks** your folder recursively (skips hidden files/folders)
2. **Pre-filters** by file size — only files sharing the same size could be duplicates
3. **SHA-256 hashes** the content of those candidates (fast: skips unique-size files entirely)
4. **Groups** files with matching hashes and presents them
5. **Moves** your chosen files to Trash (safe — nothing is permanently deleted)

---

## Interactive UI (CLI)

For each duplicate group you'll see all copies with their:
- Filename
- Full path  
- Size
- Last modified date

Then choose which copy to **keep** (others go to Trash), or **skip** the group.

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Group 1 of 4  (identical audio content)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  [1]  kick_drum_01.wav
       /Users/you/Samples/drums/kick_drum_01.wav
       Size: 1.2 MB  Modified: 2024-03-15 11:22

  [2]  Kick Hard v2 FINAL.wav
       /Users/you/Desktop/old stuff/Kick Hard v2 FINAL.wav
       Size: 1.2 MB  Modified: 2023-09-02 09:14

  Keep which file? (1-2)  [s]kip  [a]ll skip  [q]uit
  > 1
  ✓ Keeping: kick_drum_01.wav
  🗑  Trashed: Kick Hard v2 FINAL.wav
```

### Controls

| Input | Action |
|-------|--------|
| `1`, `2`, ... | Keep that file, trash the rest |
| `s` or Enter  | Skip this group (keep all) |
| `a`           | Skip all remaining groups |
| `q`           | Quit |

---

## Supported formats

Audio: `.wav` `.aiff` `.aif` `.mp3` `.flac` `.ogg` `.m4a` `.aac` `.opus` `.wma`

Use `--all` to scan every file type regardless of extension.

---

## Safety

- Files are moved to **Trash/Recycle Bin**, not permanently deleted
- You can restore anything before emptying the Trash
- Hidden files and `.DS_Store` are automatically ignored

**macOS:** Uses built-in `osascript`  
**Linux:** Uses `gio trash` or moves to `~/.local/share/Trash/`  
**Windows:** Uses recycle bin via Go standard library

---

## Building from Source

```bash
# Clone and build
git clone https://github.com/yourusername/dupclean.git
cd dupclean
go build -o dupclean .

# Run
./dupclean ~/Music

# Install to PATH
sudo mv dupclean /usr/local/bin/
```

### Cross-compilation

```bash
# macOS
make cross-darwin-local

# Linux  
make cross-linux

# Windows
make cross-windows

# All platforms
make release
```

---

## Contributing

Contributions are welcome! Here's how to help:

### Reporting Issues

- Found a bug? [Open an issue](https://github.com/PopolQue/dupclean/issues)
- Include steps to reproduce, expected vs actual behavior
- Include your OS, Go version, and relevant details

### Submitting Changes

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b my-feature`
3. **Make** your changes
4. **Test** with `go test ./...`
5. **Commit** with clear messages: `git commit -m "Add feature X"`
6. **Push** and submit a PR

### Code Style

- Run `go fmt` before committing
- Keep functions small and focused
- Add tests for new functionality
- Update documentation for user-facing changes

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Project Structure

```
dupclean/
├── main.go          # Entry point
├── gui/app.go       # GUI implementation (Fyne)
├── scanner/         # Core duplicate detection
│   └── scanner.go   # Hashing & file scanning
├── ui/              # Terminal UI
│   └── ui.go       # Interactive CLI prompts
└── Makefile        # Build commands
```

---

## License

MIT License — see LICENSE file for details.
