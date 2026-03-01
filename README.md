# 🎧 DupClean — Audio Duplicate Cleaner

A fast, content-aware duplicate file scanner for music producers and DJs.

**The key insight:** Two files can have completely different names but identical audio data.  
DupClean hashes file *contents* — not names — so it catches every duplicate no matter how it was renamed.

---

## Install

```bash
# Clone or download this repo, then:
go build -o dupclean .
sudo mv dupclean /usr/local/bin/
```

Or just run directly:
```bash
go run . ~/Music/Samples
```

## Usage

```bash
# Scan a folder for duplicate audio files
dupclean ~/Music/Samples

# Also include non-audio files
dupclean ~/Music/Samples --all
```

## How it works

1. **Walks** your folder recursively (skips hidden files/folders)
2. **Pre-filters** by file size — only files sharing the same size could be duplicates
3. **SHA-256 hashes** the content of those candidates (fast: skips unique-size files entirely)
4. **Groups** files with matching hashes and presents them interactively
5. **Moves** your chosen files to macOS Trash (safe — nothing is permanently deleted)

## Interactive UI

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

## Controls

| Input | Action |
|-------|--------|
| `1`, `2`, ... | Keep that file, trash the rest |
| `s` or Enter  | Skip this group (keep all) |
| `a`           | Skip all remaining groups |
| `q`           | Quit |

## Supported formats

`.wav` `.aiff` `.aif` `.mp3` `.flac` `.ogg` `.m4a` `.aac` `.opus` `.wma`

Use `--all` to scan every file type regardless of extension.

## Safety

- Files are moved to **macOS Trash**, not permanently deleted
- You can restore anything from Trash before emptying it
- Hidden files and `.DS_Store` are automatically ignored
- Requires `osascript` (built into macOS) — or optionally install `trash` via Homebrew for faster trashing

```bash
brew install trash   # optional, slightly faster
```
