# TidyBot: Daemon

**Status:** In-Development
**Author:** Gemini CLI
**Last Updated:** 2026-05-30

## 1. Overview & Scope

Background service monitoring filesystem events via FSEvents and triggering the Rule Engine.

## 2. Technical Architecture

- **Memory Management:** Minimally allocated state; utilizes streaming event buffers to prevent heap bloat.
- **Concurrency Model:** Single-threaded FSEvent consumer feeding a buffered worker channel.
- **Syscall/IO Strategy:** Direct `fsutil` wrapper for native FSEvents. Atomic file watches initialized at startup.

## 3. Interface Definition (API Contract)

```go
type Watcher interface {
    Start(ctx context.Context) error
    Stop() error
    ReloadConfig() error
}
```

## 4. Error Handling & Safety

- **Propagation:** Critical FSEvent errors cause graceful shutdown/re-init.
- **Safety Audits:** Watcher respects `dupclean` global ignore patterns.

## 5. Performance Constraints

- **Latency:** Event debounce < 2s; CPU usage < 1% at idle.

## 6. Verification Plan

- Integration tests simulating rapid file creation/modification.

---
[Link back to Master Overview](./../design.md)
