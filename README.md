# DupClean — All-in-one Disk Cleanup Tool

[![Build & Release](https://github.com/PopolQue/dupclean/actions/workflows/release.yml/badge.svg)](https://github.com/PopolQue/dupclean/actions/workflows/release.yml)
[![Tests](https://github.com/PopolQue/dupclean/actions/workflows/test.yml/badge.svg)](https://github.com/PopolQue/dupclean/actions/workflows/test.yml)
![Coverage](https://img.shields.io/badge/Coverage-0%25-red)
![Coverage (no GUI)](https://img.shields.io/badge/Coverage_(no_GUI)-0%25-red)

A fast, content-aware disk cleanup suite for music producers, DJs,
photographers, developers, and anyone with a messy hard drive.

DupClean combines an advanced duplicate finder (capable of detecting visually
similar photos), a safe system/browser cache cleaner, and an interactive disk
space analyzer into a single, unified interface.

---

## Features

- **Duplicate Finder** — Multi-mode scanning (Audio, Byte, Photo).
- **Cache Cleaner** — Safely reclaim space by clearing browser, system, and
  developer caches.
- **Disk Analyzer** — Fast, concurrent filesystem walker to visualize disk
  usage.
- **Safety** — Files moved to Trash, never permanently deleted. See
  [docs/SAFETY.md](docs/SAFETY.md).
- **GUI + CLI** — Modern graphical interface and terminal mode.

## Learn More

- [How it works](docs/HOW_IT_WORKS.md)
- [Supported formats](docs/SUPPORTED_FORMATS.md)
- [Project structure](docs/PROJECT_STRUCTURE.md)

---

## Installation

### GUI Version

Install via Homebrew (macOS):

```bash
brew install --cask PopolQue/dupclean/dupclean
```

Or download from [Releases](https://github.com/PopolQue/dupclean/releases).

### CLI Version

Download the binary for your platform from
[Releases](https://github.com/PopolQue/dupclean/releases).

---

## Usage

### GUI Mode

Launch `dupclean` or from your Applications folder.

### CLI Mode

#### Duplicate Finder

```bash
dupclean ~/Music                      # Audio mode (default)
dupclean ~/Documents --mode=byte      # Byte mode
dupclean ~/Photos --mode=photo        # Photo mode
```

#### Disk Analyzer

```bash
dupclean analyze ~/Downloads
```

#### Cache Cleaner

```bash
dupclean clean
```

See [docs/CLI_UI.md](docs/CLI_UI.md) for interactive terminal controls.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and
guidelines.

---

## License

MIT License — see [LICENSE](LICENSE) file for details.
