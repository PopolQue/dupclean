# DupClean Project Guidelines

## Engineering Standards

- **Testing:** ALWAYS run `go test ./...` and update related tests after making a code change.
  You must add a new test case to the existing test file (if one exists) or create a new test
  file to verify your changes. **Priority:** The `scanner` logic is the highest-value target due
  to safety implications (incorrect identification could cause users to trash the wrong file).
  Solid scanner tests are worth the investment even if GUI coverage stays low.
  - Write **table-driven tests** using `[]struct{ name, ... }` slices and `t.Run(tc.name, ...)`.
  - Use `t.Parallel()` in unit tests unless shared mutable state prevents it.
  - Test **error paths** as thoroughly as happy paths; scanner correctness requires it.
- **Linting:** ALWAYS run the linter (e.g., `golangci-lint run`) before committing. The linter
  is the first line of defence against goroutine leaks (`bodyclose`, `govet`, `staticcheck`).
- **Types & Warnings:** NEVER use hacks like disabling or suppressing warnings, bypassing the
  type system, or employing "hidden" logic.
- **Versioning:** When bumping the version, always follow the pattern
  `0.(Big Changes).(Small Changes).(Bugfixes)`. Only bump when preparing a new release.
- **Releasing:** Bump the version (in `internal/version/version.go`, `gui/changelog.go`,
  `dupclean.rb`, and `Casks/dupclean.rb`), commit, tag (`git tag v0.x.y.z`), and push both.
- **CodeQL CI:** Uses `.github/workflows/codeql.yml` for CGO. Keep GitHub's "Default" Code
  Scanning setup disabled to prevent conflicts.

---

## Go Idioms — Required Patterns

These are **non-negotiable defaults**. Deviate only with an explicit comment explaining why.

### Error handling

- Return errors; never swallow them silently (`_ = someErr` is forbidden except in `defer`
  cleanup where the error truly cannot be acted on — add a comment).
- Wrap errors with context: `fmt.Errorf("scanning %s: %w", path, err)`.
- Use `errors.Is` / `errors.As` for inspection. Never type-assert on `error` directly.
- Sentinel errors live in the package that owns the concept, named `ErrXxx`.

### Interfaces & structs

- Accept interfaces, return concrete types (standard Go advice).
- Keep interfaces small — one or two methods. Prefer composing small interfaces.
- Unexported fields; provide constructor functions (`NewScanner(...)`) that validate inputs.

### Naming

- Acronyms are all-caps: `ID`, `URL`, `HTTP` — never `Id`, `Url`, `Http`.
- Receivers are short (one or two letters), consistent per type. Never `self` or `this`.
- Package names are lowercase, single words, no underscores.

### Resource management

- Every `os.Open` / `os.Create` / `http.Response` must have a paired `defer f.Close()` in
  the **same function scope** as the open call.
- Check the error returned by `Close()` on writable files:

```go
  if err := f.Close(); err != nil {
      return fmt.Errorf("closing %s: %w", path, err)
  }
```

---

## Goroutine Discipline — Zero-Leak Policy

Goroutines are cheap to start and expensive to leak. Every goroutine launched must have a
**documented exit condition** in a comment above the `go` statement.

### Rule 1 — Always propagate context

Every function that may block or do I/O must accept a `context.Context` as its **first
parameter** and respect cancellation:

```go
func (s *Scanner) Walk(ctx context.Context, root string) error {
    return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if ctx.Err() != nil {
            return ctx.Err() // abort the walk on cancellation
        }
        // ...
    })
}
```

Never store a context in a struct field.

### Rule 2 — Worker pools must drain completely

Use the producer/worker/closer pattern. The producer closes the channel; workers drain it;
a `sync.WaitGroup` guarantees all workers have exited before the caller continues:

```go
jobs := make(chan string, bufferSize)
var wg sync.WaitGroup

for range numWorkers {
    wg.Add(1)
    go func() {           // exits when jobs is closed and drained
        defer wg.Done()
        for path := range jobs {
            if ctx.Err() != nil {
                return
            }
            process(path)
        }
    }()
}

for _, p := range paths {
    jobs <- p
}
close(jobs) // signal workers to exit
wg.Wait()   // block until every worker has returned
```

### Rule 3 — errgroup over raw goroutines + WaitGroup for error-bearing work

Prefer `golang.org/x/sync/errgroup` when workers can fail:

```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return scanDir(ctx, dir) })
if err := g.Wait(); err != nil { ... }
```

### Rule 4 — Never leak on early return

If a goroutine writes to a channel and the reader may return early (e.g., on error), the
channel must be **buffered** at least as large as the number of sends, *or* the writer must
select on `ctx.Done()`:

```go
// WRONG — blocks forever if receiver returns early
results <- value

// CORRECT — unblocks on cancellation
select {
case results <- value:
case <-ctx.Done():
    return ctx.Err()
}
```

### Rule 5 — Timer and ticker hygiene

Always `defer t.Stop()` / `defer ticker.Stop()` immediately after creation. Always drain
the channel after Stop when you need to guarantee no stale tick is processed.

### Rule 6 — Fyne callbacks run on the GUI thread; never block them

Any slow work triggered by a Fyne callback (button press, list selection) must be offloaded:

```go
btn.OnTapped = func() {
    go func() { // exits when the scan completes or ctx is cancelled
        if err := s.Run(ctx); err != nil {
            // update UI via binding or fyne thread-safe helpers
        }
    }()
}
```

Never call `time.Sleep`, file I/O, or network calls directly inside a Fyne callback.

---

## Local Reference Libraries

The following directories contain library source code for local reference (gitignored):

- **/fyne/**: Full Fyne toolkit source code.
  - `widget/` — built-in widgets
  - `layout/` — standard layouts
  - `theme/` — theme definitions and icons
  - `dialog/` — standard dialogs
  - `container/` — Tabs, Split, Scroll, etc.
  - `canvas/` — low-level drawing primitives

## UI/UX Conventions

- **Section Headers:** Use `createSectionHeader(title, subtitle string)` in `gui/app.go`.
- **Theme API:** Always use `theme.Color(theme.ColorNamePrimary)`, never deprecated methods
  like `theme.PrimaryColor()`.
- **Dialog Size:** Preferred minimum size for info/changelog dialogs is `500×400`.
