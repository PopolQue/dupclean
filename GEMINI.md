# DupClean Project Guidelines

## Engineering Standards

- **Testing:** ALWAYS run `go test ./...` and update related tests after making a code change. You must add a new test case to the existing test file (if one exists) or create a new test file to verify your changes.
- **Linting:** ALWAYS run the linter (e.g., `golangci-lint run`) before committing to ensure code quality.
- **Types & Warnings:** NEVER use hacks like disabling or suppressing warnings, bypassing the type system, or employing "hidden" logic.
- **Versioning:** Always follow the pattern `0.(Big Changes).(Small Changes).(Bugfixes)`. Before committing, ensure the version has been bumped appropriately.
- **Releasing:** After committing a version bump (in `internal/version/version.go` and `gui/changelog.go`), always tag the commit with the version (e.g., `git tag v0.x.y.z`) and push both the commit and the tag.
- **CodeQL CI:** The project uses a custom `.github/workflows/codeql.yml` to support CGO. Ensure that GitHub's "Default" Code Scanning setup is disabled in the repository security settings to prevent processing conflicts.

## Local Reference Libraries

The following directories contain library source code for local reference (gitignored). Use these to verify implementation details, available APIs, and internal behaviors:

- **/fyne/**: Full Fyne toolkit source code.
  - `widget/`: Built-in widgets (Button, Entry, List, etc.)
  - `layout/`: Standard layouts (Box, Grid, Border, etc.)
  - `theme/`: Theme definitions and standard icons.
  - `dialog/`: Standard dialog types.
  - `container/`: Specialized containers (Tabs, Split, Scroll, etc.)
  - `canvas/`: Low-level drawing primitives (Text, Image, Rectangle).

## UI/UX Conventions

- **Section Headers:** Use the shared `createSectionHeader(title, subtitle string)` in `gui/app.go` for all tool views.
- **Theme API:** Always use modern theme color access: `theme.Color(theme.ColorNamePrimary)` instead of deprecated methods like `theme.PrimaryColor()`.
- **Dialog Size:** Preferred minimum size for info/changelog dialogs is `500x400`.
