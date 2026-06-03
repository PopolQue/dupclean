# TidyBot: Automated File Triage

**Status:** In-Development
**Author:** Gemini CLI
**Last Updated:** 2026-05-30

## 1. Overview & Scope

Background janitor for cleaning up transient files (logs, caches) based on lifecycle policies.

## 2. Technical Architecture

- **Memory Management:** Lightweight scan of file metadata (stat, not full file read).
- **Concurrency Model:** Independent background worker task; periodic interval execution.
- **Syscall/IO Strategy:** Standard `os.Stat` for age evaluation.

## 3. Interface Definition (API Contract)

```go
type TriageRule struct {
    Pattern string
    MaxAge  time.Duration
    Action  ActionType
}
```

## 4. Error Handling & Safety

- **Propagation:** Log errors for files failing triage; continue to next file.
- **Safety Audits:** Restricted to specific target directories to prevent system file corruption.

## 5. Performance Constraints

- **Latency:** Low-priority background task; throttled IO usage.

## 6. Verification Plan

- Integration tests simulating file age progression.

---
[Link back to Master Overview](./../design.md)
