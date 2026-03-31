# DupClean — Redundant File Cleaner

[![Build & Release](https://github.com/PopolQue/dupclean/actions/workflows/release.yml/badge.svg)](https://github.com/PopolQue/dupclean/actions/workflows/release.yml) [![Tests](https://github.com/PopolQue/dupclean/actions/workflows/test.yml/badge.svg)](https://github.com/PopolQue/dupclean/actions/workflows/test.yml)

A fast, content-aware duplicate file scanner for music producers, DJs, photographers, and anyone with a messy hard drive.

**The key insight:** Two files can have completely different names but identical (or similar) content. DupClean uses advanced hashing algorithms to catch every duplicate — and even finds similar photos!

---

## Features

### Core Features
- **Multi-mode scanning** — Audio, general files, or photos with similarity detection
- **Content-aware detection** — Finds duplicates by hash, not filename
- **Audio preview** — Listen to files before deciding which to keep
- **Photo similarity** — Finds resized, re-encoded, or lightly edited photos
- **Ignore rules** — Exclude specific folders or file extensions per scan
- **Safe deletion** — Files are moved to Trash, never permanently deleted
- **Cross-platform** — macOS, Windows, and Linux
- **GUI + CLI** — Graphical interface for everyday use, terminal mode for power users

### Scanner Modes

| Mode | Description | Best For |
|------|-------------|----------|
| **audio** | SHA-256 hashing for audio files | Music producers, DJs |
| **byte** | SHA-256 hashing for all file types | Documents, archives, any files |
| **photo** | Perceptual hashing (pHash) | Photographers, image collections |

---

## Installation

### Full Version (with GUI)

**macOS**
1. Download `dupclean.dmg` from [Releases](https://github.com/PopolQue/dupclean/releases)
2. Double-click to mount
3. Drag `DupClean.app` to Applications

If your Mac warns you that the file is damaged, try removing the quarantine after installing:
```bash
sudo xattr -d com.apple.quarantine /Applications/DupClean.app
```

**Windows**
1. Download `dupclean-windows-amd64.zip` from [Releases](https://github.com/PopolQue/dupclean/releases)
2. Extract and run `dupclean.exe`

**Linux**
```bash
tar -xzf dupclean-linux-amd64.tar.gz
sudo mv dupclean /usr/local/bin/
```

**Required dependencies:**
- GTK (usually pre-installed on most distros)
- `aplay` for audio preview (part of `alsa-utils`)
- On headless servers: X11 forwarding or a desktop environment

---

### CLI Version only (coming soon)

**macOS & Linux**
```bash
coming soon
```

---

## Usage

### GUI Mode

**Launch GUI:**
```bash
dupclean
# or explicitly
dupclean --gui
```

**Workflow:**
1. Select a folder to scan
2. Choose scan mode: Audio, All Files, or Photos
3. Click **Start Scan** — an ignore rules dialog appears before scanning begins
4. Review duplicate groups — preview files with ▶, delete with 🗑, or use **Keep #1 & Delete Others**
5. When done, a summary shows how many files were trashed and how much space was freed

---

### CLI Mode

**Basic scan (audio mode):**
```bash
dupclean ~/Music
```

**Scan all file types:**
```bash
dupclean ~/Documents --mode=byte
# or legacy alias:
dupclean ~/Documents --all
```

**Find similar photos:**
```bash
dupclean ~/Photos --mode=photo
dupclean ~/Photos --mode=photo --similarity=85  # Less strict (finds more matches)
```

**Show help:**
```bash
dupclean --help
```

### CLI Options

| Option | Description |
|--------|-------------|
| `--mode=<mode>` | Scanner mode: `audio` (default), `byte`, `photo` |
| `--all` | Scan all file types (same as `--mode=byte`) |
| `--similarity=<pct>` | Minimum similarity for photo mode (0-100, default: 90) |
| `--gui` | Launch GUI (not available in Homebrew build) |
| `--help` | Show help |

### Scan Modes Explained

**Audio Mode (`--mode=audio`)**
- Scans only audio files (.wav, .mp3, .flac, .aac, .ogg, etc.)
- Uses 4-stage detection: size → partial hash → full hash → byte comparison
- Fastest mode for music libraries

**Byte Mode (`--mode=byte`)**
- Scans ALL file types
- Finds exact byte-for-byte duplicates
- Perfect for documents, archives, mixed file collections

**Photo Mode (`--mode=photo`)**
- Scans image files (.jpg, .png, .gif, .webp, .bmp, .tiff)
- Uses perceptual hashing to find SIMILAR (not just identical) photos
- Catches: resized images, re-encoded JPEGs, lightly edited photos
- Adjust sensitivity with `--similarity` (default: 90%)

---

## How It Works

### Audio & Byte Mode (4-Stage Algorithm)

1. **Size Pre-Filter** — Groups files by size (instant, skips 99% of non-duplicates)
2. **Partial Hash** — Hashes first 8KB of potential matches (very fast)
3. **Full SHA-256 Hash** — Hashes entire file content for exact matches
4. **Byte Comparison** — Final verification to guarantee 100% accuracy

**Performance:** Up to **100x faster** than naive hashing because:
- Files with unique sizes are never hashed
- Files with different content at the start are rejected after 8KB
- Only likely duplicates undergo full hashing and verification

### Photo Mode (Perceptual Hashing)

1. **Decode Image** — Load and normalize the image
2. **Perceptual Hash** — Compute a 64-bit fingerprint based on image structure
3. **Hamming Distance** — Compare hashes to find similar images
4. **Group by Similarity** — Cluster images above similarity threshold

**What it catches:**
- ✅ Resized images
- ✅ Re-encoded at different quality
- ✅ Slight color adjustments
- ✅ Cropped versions
- ❌ Heavily edited or composite images

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

**Controls:**

| Input | Action |
|-------|--------|
| `1`, `2`, ... | Keep that file, trash the rest |
| `s` or Enter  | Skip this group |
| `a`           | Skip all remaining groups |
| `q`           | Quit |

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

## Supported Formats

### Audio Formats
`.wav` `.aiff` `.aif` `.mp3` `.flac` `.ogg` `.m4a` `.aac` `.opus` `.wma`

### Photo Formats
`.jpg` `.jpeg` `.png` `.gif` `.webp` `.bmp` `.tiff` `.tif`

### All File Types (Byte Mode)
Any file type — documents, archives, videos, executables, etc.

---

## Safety

- Files are moved to **Trash / Recycle Bin**, never permanently deleted
- Restore anything from Trash before emptying it
- Hidden files and `.DS_Store` are automatically ignored
- Preview files before deleting to avoid mistakes

---

## Building from Source

**Prerequisites:**
- Go 1.22 or later
- Git

**Basic build:**
```bash
git clone https://github.com/PopolQue/dupclean.git
cd dupclean
go build -o dupclean .
./dupclean
```

**Build with GUI:**
```bash
go build -tags gui -o dupclean .
./dupclean
```

**Build CLI-only (for Homebrew):**
```bash
CGO_ENABLED=0 go build -o dupclean .
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
├── main.go              # CLI entry point and argument parsing
├── gui/
│   └── app.go           # GUI implementation (Fyne)
├── scanner/
│   ├── types.go         # Scanner interface, Options, data models
│   ├── registry.go      # Scanner registry and mode management
│   ├── audio.go         # AudioScanner (SHA-256 for audio files)
│   ├── byte.go          # ByteScanner (SHA-256 for all files)
│   ├── photo.go         # PhotoScanner (perceptual hashing)
│   └── utils.go         # Shared utilities (hashing, comparison)
├── ui/
│   └── ui.go            # Interactive terminal UI
└── Makefile             # Build commands
```

---

## Contributing

Contributions are welcome!

### Development Setup

```bash
git clone https://github.com/PopolQue/dupclean.git
cd dupclean
go build -o dupclean .
./dupclean
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run specific scanner tests
go test ./scanner -v -run TestFindDuplicates
```

### Guidelines

- Found a bug? [Open an issue](https://github.com/PopolQue/dupclean/issues)
- Want to add a feature? Fork, branch, and submit a PR
- Run `go fmt ./...` before committing
- Run `go test ./...` to verify nothing is broken
- Follow existing code style and conventions

---

## License

MIT License — see [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [goimagehash](https://github.com/corona10/goimagehash) for perceptual hashing
- [Fyne](https://fyne.io/) for the cross-platform GUI framework
- Homebrew community for package distribution
