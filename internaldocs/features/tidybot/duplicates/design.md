# TidyBot: Intelligent Duplicate Handling

**Status:** In-Development
**Author:** Gemini CLI
**Last Updated:** 2026-05-30

## 1. Overview & Scope

Automated duplicate cleanup based on metadata policies, bridging the Scanner and Engine.

## 2. Technical Architecture

- **Memory Management:** Efficient iteration over duplicate groups; batch processing of deletion commands.
- **Concurrency Model:** Sequential execution of deletion actions to guarantee atomicity.
- **Syscall/IO Strategy:** Atomic file removal operations or moves to OS trash.

## 3. Interface Definition (API Contract)

```go
type DuplicatePolicy interface {
    ShouldKeep(f1, f2 FileInfo) bool
    ApplyAction(action ActionType, f FileInfo) error
}
```

## 4. Error Handling & Safety

- **Propagation:** Immediate halt of deletion if file permission error encountered.
- **Safety Audits:** "Dry-run" mode is MANDATORY; requires explicit user override.

## 5. Performance Constraints

- **Latency:** Dependent on filesystem throughput; batch cleanup operations.

## 6. Verification Plan

- E2E testing using simulated duplicate groups in a sandboxed environment.

---
[Link back to Master Overview](./../design.md)
