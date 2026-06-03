# TidyBot: Rule Engine

**Status:** In-Development **Author:** Gemini CLI **Last Updated:** 2026-05-30

## 1. Overview

The execution core. Responsible for applying the parsed `Rule` ASTs against file
system events provided by the [Daemon](./../daemon/design.md).

## 2. API Contract (Go)

```go
type Engine struct {
    mu    sync.RWMutex
    rules []Rule
}

func (e *Engine) ProcessFile(ctx context.Context, path string) error {
    // 1. Thread-safe Rule lookup
    // 2. Condition evaluation
    // 3. Action execution
}
```

## 3. Concurrency & State Model

- **Locking:** `sync.RWMutex` protects the `rules` slice. Rule evaluation is
  read-locked.
- **Worker Pool:** File events are processed via a buffered channel, handled by
  a worker pool to prevent UI blocking.

## 4. Error Handling & Recovery

- **Policy:** Atomic actions. If an action fails (e.g., move failed), log error
  and halt rule chain for that file.
- **Retries:** Exponential backoff for transient IO errors (e.g., file busy).

## 5. Metadata Processing

- **Implementation:** Uses `internal/fsutil` for attribute fetching.
- **Caching:** Per-file metadata results cached for 5s to avoid excessive
  syscalls during rule evaluation.

## 6. Testing Strategy

- Mocking the file system interface for deterministic condition testing.

---

[Link back to Master Overview](./../design.md)
