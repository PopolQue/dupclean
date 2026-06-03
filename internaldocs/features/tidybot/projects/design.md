# TidyBot: Project-Based Asset Management

**Status:** In-Development
**Author:** Gemini CLI
**Last Updated:** 2026-05-30

## 1. Overview & Scope

Automated sorting of incoming assets into structured project directories based on rules.

## 2. Technical Architecture

- **Memory Management:** Efficient path matching using optimized trie/glob algorithms.
- **Concurrency Model:** Parallel file processing within dedicated watch-directories.
- **Syscall/IO Strategy:** Standard filesystem movement commands; atomic renames.

## 3. Interface Definition (API Contract)

```go
type AssetMatcher interface {
    Match(filename string) (Project, error)
}
```

## 4. Error Handling & Safety

- **Propagation:** Error if destination project structure is invalid; halt move.
- **Safety Audits:** Validation of project paths before movement.

## 5. Performance Constraints

- **Latency:** Near-instant reaction to file creation events.

## 6. Verification Plan

- Unit tests for directory-based matching logic.

---
[Link back to Master Overview](./../design.md)
