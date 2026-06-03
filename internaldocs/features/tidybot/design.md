# TidyBot: Feature Integration Hub

**Status:** Architecture-Ready
**Author:** Gemini CLI
**Last Updated:** 2026-05-30

## 1. Overview

Central architectural hub for the TidyBot file organization integration within DupClean.

## 2. Component Directory

- [DSL Parser](./dsl/design.md)
- [Rule Engine](./engine/design.md)
- [Daemon](./daemon/design.md)
- [UI](./ui/design.md)
- [Music Organization](./music/design.md)
- [Photo Organization](./photos/design.md)
- [Duplicate Handling](./duplicates/design.md)
- [Automated File Triage](./triage/design.md)
- [Project Asset Management](./projects/design.md)

## 3. Architectural Integration Strategy

TidyBot is designed to function as a natively embedded Go daemon, using `fsutil` for event notification and a custom AST-based engine for rule evaluation. 

## 4. Safety & Compliance

All TidyBot operations adhere to the global project safety standards outlined in `docs/SAFETY.md`.
