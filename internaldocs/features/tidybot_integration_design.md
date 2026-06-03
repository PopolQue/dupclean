# TidyBot Integration Design Document

**Status:** Draft **Date:** 2026-05-30 **Component:** TidyBot Feature
Integration

## 1. Overview

This document outlines the architectural and implementation approach for
integrating TidyBot's automated file organization capabilities into the existing
`dupclean` codebase.

TidyBot, originally designed as a daemon, will be implemented natively within
`dupclean` to allow for rule-based, background file organization alongside
existing duplicate cleaning functionality.

## 2. Goals & Non-Goals

### 2.1 Goals

- **Automated Organization:** Provide a mechanism for zero-friction file
  organization running in the background.
- **Native Go DSL:** Implement a robust, Go-based DSL parser and evaluator for
  TidyBot rules.
- **Integration:** Deeply integrate rule engine, watcher daemon, and GUI
  components within `dupclean`.
- **Safety:** Ensure file manipulation is reversible/safe, maintaining
  `dupclean`'s high safety standards.
- **Native macOS Integration:** Utilize `FSEvents` and `launchd` for seamless
  daemon operations on macOS.

### 2.2 Non-Goals

- Cross-platform support for v1 (macOS only initially).
- Cloud synchronization.
- Complex content-based file inspection (OCR/deep metadata).

## 3. Architecture

The integration follows a layered approach in Go:

```text
internal/tidybot/
├── dsl/          # Tokenizer, Parser, AST definition
├── engine/       # Rule evaluator & Action executor
├── daemon/       # FSEvents watchdog, debouncer, runner
└── ui/           # GUI components for rule management
```

### 3.1 Data Flow

```text
File System Event (FSEvent)
  → Daemon Watcher (Debounce 2s)
    → Engine.Process(path)
      → EvaluateRule(path)
        → All Conditions Match?
          → Execute Actions
            → ActionResult (Log/UI Update)
```

## 4. The Rule DSL

The DSL will mirror the English-like syntax defined in the original TidyBot
design, implemented via a custom Go parser.

### Syntax Example

```text
WHEN ext IN ["pdf", "docx"] AND size > 10MB THEN MOVE TO "~/Documents/BigFiles"
WHEN name MATCHES "Invoice*" THEN TAG "Invoice"
```

## 5. Implementation Roadmap

### Phase 1: Core DSL Engine (`internal/tidybot/dsl`, `internal/tidybot/engine`)

- Define `Rule`, `Condition`, and `Action` structs in Go.
- Implement tokenization and parsing logic.
- Develop the evaluation engine.

### Phase 2: Daemon Watcher (`internal/tidybot/daemon`)

- Utilize existing `fsutil` to implement FSEvents watcher.
- Implement rule loading and hot-reloading capability.
- Integrate with `launchd` for background service management.

### Phase 3: GUI Integration (`gui/`)

- Create Rule Management tab in the GUI.
- Implement syntax-highlighted editor for rule creation.
- Add visual rule builder.

### Phase 4: Validation & Testing

- Comprehensive unit tests for DSL parsing.
- Integration tests for daemon watching.
- E2E testing of file organization actions.

## 6. Security & Safety

- **Sandboxing:** Actions should be restricted to user-configured directories.
- **Dry-run:** All actions default to dry-run mode unless explicitly enabled.
- **Atomic Operations:** Use atomic move/copy operations to prevent data
  corruption during processing.
- **Logging:** All organization actions MUST be logged for transparency and
  potential rollback.

## 7. Testing Strategy

| Component      | Test Method             | Focus                                    |
| :------------- | :---------------------- | :--------------------------------------- |
| **DSL Parser** | Table-driven Unit Tests | Valid/Invalid syntax coverage            |
| **Engine**     | Integration Tests       | Correct evaluation of conditions/actions |
| **Daemon**     | Integration Tests       | Event debouncing, watcher stability      |
| **GUI**        | Component Testing       | UI rendering and event dispatching       |

---

***End of Design Document***
