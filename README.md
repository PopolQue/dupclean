# DupClean — Redundant File Cleaner

[![Build & Release](https://github.com/PopolQue/dupclean/actions/workflows/release.yml/badge.svg)](https://github.com/PopolQue/dupclean/actions/workflows/release.yml) [![Tests](https://github.com/PopolQue/dupclean/actions/workflows/test.yml/badge.svg)](https://github.com/PopolQue/dupclean/actions/workflows/test.yml) ![Coverage](https://img.shields.io/badge/Coverage-23.7%25-red)

A fast, content-aware duplicate file scanner for music producers, DJs, and anyone with a messy hard drive.

**The key insight:** Two files can have completely different names but identical content.  
DupClean hashes file *contents* — not names — so it catches every duplicate no matter how it was renamed.

---

## Features

- **Content-aware scanning** — finds duplicates by file hash, not filename
- **Audio preview** — listen to files before deciding which to keep
- **Ignore rules** — exclude specific folders or file extensions per scan
- **Safe deletion** — files are moved to Trash, never permanently deleted
- **Cross-platform** — macOS, Windows, and Linux
- **GUI + CLI** — graphical interface for everyday use, terminal mode for power users

---

## Installation

### macOS

**Option 1: DMG (Recommended)**
1. Download `dupclean.dmg` from [Releases](https://github.com/PopolQue/dupclean/releases)
2. Double-click to mount
3. Drag `DupClean.app` to Applications

**Option 2: Binary**
```bash
tar -xzf dupclean-darwin-arm64.tar.gz
sudo mv dupclean /usr/local/bin/
```

**Option 3: Homebrew**
```bash
# Coming soon
```

---

### Linux

```bash
tar -xzf dupclean-linux-amd64.tar.gz
sudo mv dupclean /usr/local/bin/
```

**Required dependencies:**
- GTK (usually pre-installed on most distros)
- `aplay` for audio preview (part of `alsa-utils`)
- On headless servers: X11 forwarding or a desktop environment

---

### Windows

Download and extract the `.zip` from [Releases](https://github.com/PopolQue/dupclean/releases), then run `dupclean.exe` directly or double-click in File Explorer.

---

## Usage

### GUI Mode

Launch without arguments to open the graphical interface:

```bash
dupclean
# or explicitly
dupclean --gui
```

**Workflow:**
1. Select a folder to scan
2. Optionally check "Scan all file types" (default: audio only)
3. Click **Start Scan** — an ignore rules dialog appears before scanning begins
4. Review duplicate groups — preview files with ▶, delete with 🗑, or use **Keep #1 & Delete Others**
5. When done, a summary shows how many files were trashed and how much space was freed

---

### CLI Mode

```bash
# Scan a folder for duplicate audio files
dupclean ~/Music/Samples

# Also scan non-audio files
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
5. **Keeps** your chosen file and moves to duplicates to Trash (safe — nothing is permanently deleted)

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
=======
**Controls:**

| Input | Action |
|-------|--------|
| `1`, `2`, ... | Keep that file, trash the rest |
| `s` or Enter  | Skip this group |
| `a`           | Skip all remaining groups |
| `q`           | Quit |

---

## How It Works

1. **Walks** the folder recursively, skipping hidden files and ignored paths
2. **Pre-filters** by file size — only same-size files can be duplicates
3. **SHA-256 hashes** the content of candidates (skips unique-size files entirely for speed)
4. **Groups** files with matching hashes and presents them for review
5. **Moves** chosen files to Trash — nothing is permanently deleted

---

## Ignore Rules

Before each scan, a dialog lets you configure rules for that session:

- **Folders to ignore** — use the folder picker to exclude specific directories (e.g. a backup folder you want to keep duplicates in)
- **Extensions to ignore** — comma-separated list (e.g. `.txt, .pdf`) to skip certain file types

Ignore rules reset after each scan and are not saved between sessions.

---

## Audio Preview

Click the ▶ button on any file in a duplicate group to preview it. Starting a new preview automatically stops the previous one. Trashing a file also stops playback if that file is currently playing.

**Preview uses native OS audio playback:**
- macOS: `afplay` (built-in)
- Linux: `aplay` (install via `sudo apt install alsa-utils`)
- Windows: PowerShell `Media.SoundPlayer`

---

## Supported Audio Formats

`.wav` `.aiff` `.aif` `.mp3` `.flac` `.ogg` `.m4a` `.aac` `.opus` `.wma`

Use `--all` in CLI mode or check "Scan all file types" in the GUI to scan every file type regardless of extension.

---

## Safety

- Files are moved to **Trash / Recycle Bin**, never permanently deleted
- Restore anything from Trash before emptying it
- Hidden files and `.DS_Store` are automatically ignored

---

## Building from Source

```bash
git clone https://github.com/PopolQue/dupclean.git
cd dupclean
go build -o dupclean .
./dupclean
```

### Cross-compilation

```bash
make cross-darwin-local   # macOS
make cross-linux          # Linux
make cross-windows        # Windows
make release              # All platforms
```

---

## Project Structure

```
dupclean/
├── main.go              # Entry point, CLI arg handling
├── gui/
│   └── app.go           # GUI implementation (Fyne)
├── scanner/
│   └── scanner.go       # Content hashing & duplicate detection
├── ui/
│   └── ui.go            # Interactive terminal UI
└── Makefile             # Build commands
```

---

## Contributing

Contributions are welcome.

- Found a bug? [Open an issue](https://github.com/PopolQue/dupclean/issues)
- Want to add a feature? Fork, branch, and submit a PR
- Run `go fmt ./...` before committing
- Run `go test ./...` to verify nothing is broken

---

## License

MIT License — see LICENSE file for details.
