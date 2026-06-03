# TidyBot: UI

**Status:** In-Development
**Author:** Gemini CLI
**Last Updated:** 2026-05-30

## 1. Overview & Scope

Graphical management interface for TidyBot rules and configurations, implemented in Fyne.

## 2. Technical Architecture

- **Memory Management:** Efficient widget garbage collection by Fyne runtime.
- **Concurrency Model:** UI events (clicks/inputs) operate on the main loop; rule evaluation/reloading via background goroutines to prevent hangs.
- **Syscall/IO Strategy:** Direct reading of `.tidybot` config files on load/save operations.

## 3. Interface Definition (API Contract)

```go
type RuleEditor interface {
    LoadRules() ([]Rule, error)
    SaveRules([]Rule) error
}
```

## 4. Error Handling & Safety

- **Propagation:** Validation of DSL syntax in the UI *before* attempted save.
- **Safety Audits:** Confirmation dialogs for "Clear All Rules" actions.

## 5. Performance Constraints

- **Latency:** Instant reactivity to user input; Rule reloads < 100ms.

## 6. Verification Plan

- Component-level unit tests for widgets; integration tests for rule load/save lifecycle.

---
[Link back to Master Overview](./../design.md)
