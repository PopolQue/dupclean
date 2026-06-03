# TidyBot — Design Document

**Version:** 1.0  
**Date:** 2026-05-30  
**Status:** Draft

---

## 1. Overview

TidyBot is a macOS background daemon that watches user-configured directories
and automatically organises files according to a custom rule language. Rules are
written in a human-readable DSL (domain-specific language) and evaluated in real
time as files arrive. A React-based UI provides a syntax-highlighted editor,
visual rule builder, and configuration management.

The core design goal is that rules should read like plain English while
remaining precise enough to express complex multi-condition logic with variable
interpolation, chained actions, and shell script integration.

---

## 2. Goals & Non-Goals

### Goals

- Zero-friction file organisation that runs invisibly in the background
- A rule language expressive enough to replace manual sorting entirely
- Hot-reload of rules without restarting the daemon
- Native macOS integration: FSEvents, Finder tags, `osascript` notifications,
  launchd
- Dry-run mode for safe experimentation
- A config UI that lowers the barrier to writing DSL rules

### Non-Goals

- Cross-platform support (Linux/Windows are out of scope for v1)
- Cloud sync or remote rule management
- File content inspection (OCR, metadata parsing beyond filesystem attributes)
- Recursive directory watching

---

## 3. Architecture

```filetree
tidybot/
├── daemon/
│   ├── dsl.py        # Tokeniser, parser, AST
│   ├── engine.py     # Rule evaluator + action executor
│   └── daemon.py     # Watchdog loop, launchd integration, hot-reload
└── ui/
    └── src/
        └── App.jsx   # Config UI: editor, visual builder, vars reference
```

### Data flow

```flow
File event (FSEvent)
  → TidyHandler (debounce 2s)
    → RuleEngine.process_file(path)
      → evaluate_rule() for each Rule
        → evaluate_condition() for each Condition   [all must pass]
          → execute_action() for each Action
            → ActionResult
```

### Component responsibilities

| Component   | Responsibility                                                         |
| ----------- | ---------------------------------------------------------------------- |
| `dsl.py`    | Tokenise and parse `.tidybot` files into typed `Rule` objects          |
| `engine.py` | Evaluate conditions against `pathlib.Path`, execute filesystem actions |
| `daemon.py` | FSEvents via watchdog, debounce, hot-reload, launchd plist management  |
| `App.jsx`   | Syntax-highlighted DSL editor, visual builder, watch-dir management    |

---

## 4. The Rule DSL

### 4.1 Syntax

```DSL
WHEN <condition> [AND <condition> ...] THEN <action> [AND <action> ...]
```

Lines beginning with `#` are comments. Multi-line rules can use trailing `\`
continuation.

### 4.2 Conditions

| Field  | Operators         | Value format   | Example                   |
| ------ | ----------------- | -------------- | ------------------------- |
| `ext`  | `IS`, `IN`        | string or list | `ext IN ["pdf", "docx"]`  |
| `name` | `IS`, `MATCHES`   | string or glob | `name MATCHES "Invoice*"` |
| `size` | `>` `<` `>=` `<=` | number + unit  | `size > 500MB`            |
| `age`  | `>` `<` `>=` `<=` | number + unit  | `age > 30d`               |
| `dir`  | `IS`              | path string    | `dir IS "~/Downloads"`    |
| `tag`  | `IS`, `IN`        | string or list | `tag IS "Work"`           |

Size units: `B`, `KB`, `MB`, `GB`  
Age units: `h` (hours), `d` (days), `w` (weeks), `m` (months)

Multiple conditions on one rule use AND semantics — all must pass.

### 4.3 Actions

| Verb        | Syntax                            | Notes                                       |
| ----------- | --------------------------------- | ------------------------------------------- |
| `MOVE TO`   | `MOVE TO "~/Docs/{year}"`         | Creates destination directories as needed   |
| `COPY TO`   | `COPY TO "~/Backup/{ext}"`        | Preserves original                          |
| `TRASH`     | `TRASH`                           | Moves to `~/.Trash`                         |
| `RENAME TO` | `RENAME TO "{name}_{date}.{ext}"` | Stays in same directory                     |
| `NOTIFY`    | `NOTIFY "Got {filename}"`         | macOS notification via osascript            |
| `RUN`       | `RUN "~/scripts/x.sh {filepath}"` | Arbitrary shell command                     |
| `SKIP`      | `SKIP`                            | Stop processing further rules for this file |

Multiple actions on one rule are chained with AND and execute in order. If a
`MOVE` or `TRASH` action succeeds, subsequent rules for that file are skipped
(the path no longer exists at its original location).

### 4.4 Path variables

Variables are interpolated into path templates and action arguments at execution
time.

| Variable      | Value                                                     |
| ------------- | --------------------------------------------------------- |
| `{ext}`       | Extension without dot, lowercased                         |
| `{name}`      | Filename without extension                                |
| `{filename}`  | Full filename with extension                              |
| `{filepath}`  | Absolute path to the file                                 |
| `{year}`      | 4-digit modification year                                 |
| `{month}`     | Zero-padded month (01–12)                                 |
| `{day}`       | Zero-padded day (01–31)                                   |
| `{date}`      | Full date as YYYY-MM-DD                                   |
| `{size_tier}` | `small` (<1 MB), `medium` (<100 MB), or `large` (≥100 MB) |

### 4.5 Example ruleset

```tidybot
# Sort images from Downloads into dated folders
WHEN ext IN ["jpg", "jpeg", "png", "heic"] AND dir IS "~/Downloads" \
  THEN MOVE TO "~/Pictures/Inbox/{year}/{month}"

# Archive anything over 1 GB with a notification
WHEN size > 1GB \
  THEN MOVE TO "~/Archive/{year}" AND NOTIFY "Archived large file: {filename}"

# Trash temp files older than a week
WHEN ext IN ["tmp", "crdownload", "part"] AND age > 7d THEN TRASH

# Route files by macOS Finder tag
WHEN tag IS "Work" THEN MOVE TO "~/Work/Inbox"

# Run a custom processing script on CSVs
WHEN ext IS "csv" THEN RUN "~/scripts/process_csv.sh {filepath}"
```

---

## 5. DSL Parser (`dsl.py`)

### 5.1 Pipeline

```pipeline
Raw text
  → split + join continuation lines (\)
    → parse_line() per non-blank, non-comment line
      → split at THEN boundary
        → tokenise with shlex.split() (preserves quoted strings)
          → _split_on_keyword("AND") → condition groups / action groups
            → _parse_condition() / _parse_action()
              → Condition / Action dataclass
```

### 5.2 Key types

```python
@dataclass
class Condition:
    field: str       # ext | name | size | age | dir | tag
    operator: str    # IS | IN | MATCHES | > | < | >= | <=
    value: object    # str | list[str] | int (bytes) | int (seconds)

@dataclass
class Action:
    verb: str              # MOVE | COPY | TRASH | RENAME | NOTIFY | RUN | SKIP
    argument: str | None   # path template, message, or script

@dataclass
class Rule:
    conditions: list[Condition]
    actions: list[Action]
    raw: str
    line_number: int
```

### 5.3 Error handling

`parse_ruleset()` returns `(rules, errors)` — parse failures are collected as
`DSLError` instances rather than raising, so a single bad line doesn't abort the
whole file. Errors include line number and the offending text for display in the
UI.

---

## 6. Rule Engine (`engine.py`)

### 6.1 Condition evaluation

Each condition type dispatches to its field handler:

- `ext` / `name` — string equality or `fnmatch` glob
- `size` — `Path.stat().st_size` vs. parsed byte value
- `age` — `time.time() - stat.st_mtime` vs. parsed seconds
- `dir` — `str(path.parent)` prefix match against expanded path
- `tag` — reads macOS `com.apple.metadata:_kMDItemUserTags` xattr via `xattr` +
  `plistlib`

### 6.2 Action execution

Actions are executed in order. Each returns an `ActionResult` with success flag,
message, and optional new path. If a destructive action (`MOVE`, `RENAME`,
`TRASH`) succeeds, the engine updates the tracked `filepath` variable so
subsequent actions in the same rule operate on the correct location.

Filename collisions on `MOVE` and `COPY` are resolved by appending a Unix
timestamp to the stem.

### 6.3 Dry-run mode

When `dry_run=True`, all filesystem operations are skipped. Actions log what
they would do and return success results with descriptive messages. Useful for
validating a ruleset against existing files before deploying.

---

## 7. Daemon (`daemon.py`)

### 7.1 File watching

Uses `watchdog` with the macOS FSEvents backend (kernel-level, not polling).
Each watched directory gets a non-recursive `Observer` schedule. The
`TidyHandler` debounces events with a configurable delay (default 2 s) to avoid
acting on partially-written files (e.g. downloads in progress).

### 7.2 Startup scan

When `scan_on_start: true` (default), the daemon iterates all existing files in
watched directories on launch and processes them through the rule engine. This
ensures the watched directories are in a consistent state immediately.

### 7.3 Hot-reload

The daemon checks the rules file's mtime every 5 seconds. When it changes, the
ruleset is re-parsed in place — no restart required. Parse errors are logged but
don't replace the running ruleset, so a broken edit doesn't disable the daemon.

### 7.4 launchd integration

`--install` writes a `launchd` plist to
`~/Library/LaunchAgents/com.tidybot.daemon.plist` and loads it. The plist sets
`KeepAlive: true` so the OS restarts the daemon if it crashes, and
`RunAtLoad: true` so it starts at login. `--uninstall` unloads and removes the
plist.

### 7.5 Configuration

`~/.tidybot/config.json`:

```json
{
  "watch_dirs": ["~/Downloads", "~/Desktop"],
  "rules_file": "~/.tidybot/rules.tidybot",
  "scan_on_start": true,
  "debounce_seconds": 2.0,
  "ignore_patterns": [".*", "*.part", "*.crdownload", "*.tmp"],
  "dry_run": false
}
```

---

## 8. Config UI (`App.jsx`)

The UI is a React application with four tabs.

**Editor** — Full-width text area with a syntax-highlighted overlay (keyword,
action, field, string, and number colouring). A snippet sidebar inserts common
rule templates at the cursor. A status bar shows live parse errors with line
numbers.

**Visual Builder** — Dropdown-driven form for building a single rule without
typing DSL. Condition and action rows can be added dynamically. The generated
rule is shown as highlighted DSL and inserted into the editor with one click.

**Watch Dirs** — Add and remove watched directories. Displays as pill tags.
Changes are saved to config.json and require a daemon restart.

**Variables** — Reference card for all interpolation variables and condition
fields with examples.

### 8.1 Syntax highlighting

Implemented as a transparent `<div>` overlaying the `<textarea>`, kept in sync
via scroll events. Categories:

| Token class | Colour | Examples                  |
| ----------- | ------ | ------------------------- |
| keyword     | indigo | `WHEN`, `THEN`, `AND`     |
| action      | amber  | `MOVE`, `TRASH`, `NOTIFY` |
| field       | green  | `ext`, `size`, `age`      |
| operator    | slate  | `IS`, `IN`, `MATCHES`     |
| string      | orange | `"~/Downloads"`           |
| number      | blue   | `500MB`, `30d`            |
| comment     | gray   | `# This is a comment`     |

---

## 9. File & Directory Structure

```filetree
~/.tidybot/
├── config.json          # daemon configuration
├── rules.tidybot        # rule definitions (hot-reloaded)
└── logs/
    ├── tidybot.log      # structured action log
    ├── stdout.log       # launchd stdout capture
    └── stderr.log       # launchd stderr capture

~/Library/LaunchAgents/
└── com.tidybot.daemon.plist   # installed by --install
```

---

## 10. Dependencies

| Package    | Purpose                | Install                |
| ---------- | ---------------------- | ---------------------- |
| `watchdog` | FSEvents file watching | `pip install watchdog` |
| `docx`     | (UI build only)        | `npm install -g docx`  |

No other runtime dependencies. The daemon uses only Python stdlib (`pathlib`,
`shutil`, `shlex`, `subprocess`, `signal`, `json`, `fnmatch`, `plistlib`) plus
`watchdog`.

---

## 11. Security Considerations

- The `RUN` action executes arbitrary shell commands as the current user. Rules
  should only be loaded from user-controlled files.
- The daemon operates with the user's full filesystem permissions. There is no
  sandboxing.
- The launchd service runs as the logged-in user, not root.
- Dry-run mode should always be used when evaluating an unfamiliar ruleset.

---

## 12. Future Work

- Log viewer tab in the UI (tail `tidybot.log` with action timeline)
- Rule priority / ordering controls (drag to reorder)
- "Test file" simulator — paste a path and see which rules would fire
- Conflict resolution strategies: skip, overwrite, rename-with-timestamp
  (currently hardcoded to timestamp)
- Recursive directory watching with configurable depth
- Content-based conditions: file size on disk vs apparent size, EXIF data, PDF
  metadata
- Rule import/export and community rule sharing
- Linux support via inotify (watchdog already supports it; main blocker is
  launchd-specific integration)
