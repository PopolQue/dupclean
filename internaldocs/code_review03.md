29.05.2026

dupclean is a multi-frontend (CLI + interactive TUI + Fyne GUI) duplicate file
finder, disk analyzer, and cache cleaner. It is written in Go 1.25 with a clean
modular structure (cmd/, internal/, scanner/, cleaner/, diskanalyzer/, gui/,
cli/). Maturity Level: Mid-to-Senior. The codebase shows significant engineering
maturity — proper concurrency patterns, context propagation, interface design,
multi-platform support, CI/CD with cross-compilation, and a comprehensive test
suite. However, several security, concurrency, and code quality issues prevent
it from being production-grade. Strengths: Well-separated packages, good use of
interfaces (Scanner), extensive test coverage, cross-platform awareness (inode
handling, OS-specific targets), 4-stage scanning pipeline (size grouping →
partial hash → full hash → byte verification), context cancellation throughout,
and a vendored Fyne GUI. Weaknesses: Auto-updater HTTP usage is insecure,
AppleScript command injection risk, TOCTOU race in file deletion, log file
redirection can swallow all logs, CLI entry uses goto, duplicate formatSize
across packages, and several unchecked errors. B. Critical Issues (Must Fix) B1.
Insecure HTTP Downloads in Auto-Updater (gui/updater.go)

- performUpdate() at line 281 uses http.Get(url) with no TLS verification. The
  download URL from GitHub API uses HTTPS, but there is no certificate pinning
  or signature verification on the downloaded binary.
- checkForUpdates() at line 117 creates a client with only a 10s timeout but no
  TLS config.
- Impact: Man-in-the-middle attack could serve a malicious binary that gets
  executed.
- Fix: Verify the downloaded binary's checksum (e.g., SHA-256 from the release
  assets), pin the CA cert, or use go-getter/sumcheck verification. B2.
  AppleScript Command Injection (internal/trash/trash.go + cleaner/security.go)
- moveToTrashMacOS() at line 86: script :=
  fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`,
  escapedPath) While escapeAppleScriptString() handles backslash and
  double-quote, it does not handle all AppleScript special characters (e.g., \n
  injected via filename could break the script). If a file named " \n " & do
  shell script "rm -rf ~" & " exists, this could be exploited.
- macOSInstallWithElevation() at line 401 passes user-controlled paths through
  quoted form of, which is safer but still a code-path into osascript with admin
  rights.
- SafePlayMedia() in cleaner/security.go:20 calls afplay/aplay with the path as
  a direct argument. While the path is validated to exist, there is no
  sanitization.
- Fix: Use exec.Command with separate args (already done for simple cases), but
  for AppleScript, consider using NSWorkspace via cgo or validate the path is a
  regular file with no embedded control characters. Use -e with a here-doc
  pattern rather than positional script argument. B3. TOCTOU Race in Cache
  Cleaner (gui/cache_cleaner.go)
- cleanPath() at lines 468-507 first measures the directory contents, then
  deletes via osRemoveAll(basePath) and recreates the directory. Between
  measurement and deletion, the directory could be replaced with a symlink
  pointing to an arbitrary path.
- Impact: Privilege escalation or unintended deletion of critical files.
- Fix: Measure and delete atomically. Use os.Open on the directory, then
  os.RemoveAll via the file descriptor (not the path). Or compute the size from
  the walk done during deletion. B4. Global log.SetOutput() Swallows Logs
  (gui/state.go)
- The init() function at line 44 calls log.SetOutput(logFile) to redirect all
  standard library logging to a file. This is global state — if any other
  package uses log.Printf, its output goes to the dupclean log file instead of
  stderr.
- Impact: CLI mode output capture and debugging difficulties. scanner.go in the
  cleaner package already uses log.Printf.
- Fix: Use a dedicated logger instance per package instead of the global log. Or
  only redirect in GUI mode. B5. Goroutine Leak in Photo Scanning
  (scanner/photo.go)
- In groupBySimilarity() at lines 256-318, the BK-Tree construction and Search()
  are called with the maximum Hamming distance derived from the similarity
  percentage. The formula at line 272: maxDistance := int((100 -
  s.SimilarityPct) \* 64 / 100) When SimilarityPct > 100, this becomes negative,
  making maxDistance negative, which causes Search() to add everything to the
  queue but never match anything (since distance <= maxDistance is always false
  for positive distances). Not a crash, but perplexing behavior.
- More critically, if maxDistance is very large (e.g., SimilarityPct=0 →
  maxDistance=64), the BK-tree becomes essentially a linear scan of all nodes.
  While not a leak, this negates the benefit of the tree.
- Fix: Clamp SimilarityPct to valid range (1-100) and ensure maxDistance is
  bounded. B6. Dry-Run Reports Incorrect Deletion for Directories
  (cleaner/deleter.go)
- In Delete(), lines 48-54 calculate freedBytes using entry.Size. For directory
  entries (IsDir=true), Size may be 0 or the directory's own size (typically 0
  or 4096), not its recursive contents. This underestimates dry-run savings for
  directory patterns in the cache cleaner.
- Fix: Ensure EntryInfo.Size for directories reflects recursive size, or add a
  RecursiveSize field. C. Important Improvements C1. Potential Data Race in
  deleteEntry (cleaner/deleter.go)
- The osRemoveAll = os.RemoveAll mockable variable at line 18 of mockable.go is
  a mutable global used as a function pointer. While this is for testing, if any
  test modifies it concurrently (which they do), the race detector will fire.
  C2. Unbounded Concurrency in Feeder Pattern
- diskanalyzer/walker.go:118 creates a buffered channel of size
  opts.Concurrency\*2, but the feeder goroutine walks the tree unbounded — it
  feeds all paths into the channel regardless of how many workers are
  processing. For directories with millions of files, this allocates memory for
  all paths simultaneously.
- Fix: Use a bounded semaphore or rate-limiter to prevent memory exhaustion. C3.
  Duplicate formatSize Across Packages
- cleaner/render_cli.go:271 and diskanalyzer/render_cli.go:193 both define
  identical formatSize() functions. internal/fsutil/utils.go also has
  FormatBytes() which does the same thing.
- Fix: Consolidate all size formatting into internal/fsutil and remove the
  duplicates. C4. FindDuplicates() Deprecated Path Still Used in GUI
- gui/duplicate_finder.go:29 aliases findDuplicates = scanner.FindDuplicates,
  which is marked as deprecated in scanner/scanner.go:39. The GUI should use
  ByteScanner.Scan() or AudioScanner.Scan() directly.
- Fix: Update the GUI to use the new Scanner interface directly instead of the
  deprecated wrapper. C5. No Rate Limiting on GitHub API (gui/updater.go)
- checkForUpdates() at line 117 hits the GitHub API. There is no rate limit
  handling. GitHub allows 60 unauthenticated requests/hour. If a user checks
  frequently, they'll hit the limit.
- Fix: Add caching (remember the last check time, don't re-check within 1 hour)
  and handle 403 responses gracefully. C6. renderTree() Bubble Sort
  (diskanalyzer/render_cli.go:82-88)
- Uses O(n²) bubble sort to sort directory items. For directories with thousands
  of files, this is slow.
- Fix: Use sort.Slice() like the rest of the codebase does. C7. Context
  Cancellation in Worker Pools Only Drops on ctx.Done() Read
- In scanner/engine.go:222-226 and scanner/photo.go:168-172, the worker
  goroutines check ctx.Done() only at the start of each job iteration via a
  select. If the context is cancelled while a worker is actively hashing a large
  file, the hash continues to completion.
- Fix: Use an io.LimitReader with context awareness or a separate
  context-shutdown mechanism for long-running operations. C8. protectedPaths in
  cleaner/deleter.go is Not Comprehensive
- List at line 127 includes common Unix system paths but omits macOS-specific
  paths like /System, /private, /cores, /Volumes, /Network, /home (already in
  list), /Applications. Windows paths are limited.
- Fix: Expand the list or use a more robust method (check if the path is a mount
  point or a well-known system path). C9. gui/state.go:35 Platform-Switch Uses
  filepath.Separator for OS Detection
- Uses filepath.Separator to determine if on Windows ('\\'). This works at
  runtime but is fragile and unconventional. The rest of the codebase uses
  runtime.GOOS.
- Fix: Use runtime.GOOS consistently. C10. No Gob/JSON Decoding Limits
- gui/updater.go:129 uses json.NewDecoder(resp.Body).Decode(&release) with no
  size limit on the response body. A malicious or oversized response could
  exhaust memory.
- Fix: Use io.LimitReader(resp.Body, 10*1024*1024) when decoding. C11.
  Dependency on golang.org/x/term Only Used for Terminal Width Detection
- Used in one place (diskanalyzer/render_cli.go:225) to get terminal width for
  display formatting. This is a legitimate but heavyweight dependency for a
  single function.
- Fix: Accept this as-is (the dependency is well-maintained) but consider
  vendoring only the needed function. D. Minor / Style Issues D1. Use of goto in
  cmd/root.go:88
- goto stage3 at line 88 — Go discourages goto. Replace with a boolean flag or
  function call. D2. Unused \_ Assignments
- scanner/engine.go:120: _ = width and _ = height in treemap.go:120,138 — these
  variables are assigned but unused.
- diskanalyzer/treemap.go:120,138: Same pattern, unused width/height variables.
  D3. Mixed Build Constraints
- main.go uses //go:build !gui (new style) and // +build !gui (old style). The
  old style // +build is deprecated since Go 1.17. Should only use //go:build.
- Same in all build-tag-guarded files. D4. Inconsistent Receiver Names
- scanner/photo.go:256 — groupBySimilarity uses the receiver (s *PhotoScanner).
  The naming is fine, but methods like Scan use (s *AudioScanner), (s
  *ByteScanner), (s *PhotoScanner). Consistent but could be clearer with as, bs,
  ps. D5. Error String Literals Use Redundant fmt.Sprintf
- scanner/errors.go:44,46 use fmt.Sprintf where simple string concatenation
  would suffice (or even errors.New).
- Minor, but inconsistent with other parts of the codebase. D6.
  internal/fsutil/utils.go:FormatBytes vs cleaner/render_cli.go:271 formatSize —
  Different Formatting
- FormatBytes uses "%.1f %s" while formatSize uses "%.2f %s" — they will produce
  different outputs for the same size. This is a UX inconsistency. D7.
  scanner/scanner.go Global Variables
- audioExtensions and photoExtensions are package-level map[string]bool maps.
  While these are never mutated after init (they are read-only), they violate
  the "no global mutable state" best practice. Could be var with an init() or a
  function that returns a copy. E. Concrete Recommendations E1. Security
  Hardening

1. Add checksum verification to the auto-updater (gui/updater.go). Fetch the
   release's checksums from GitHub and verify the downloaded binary before
   installation.
2. Replace AppleScript trash fallback with a native macOS solution using
   NSFileManager via cgo or the trash CLI which takes separate arguments
   (already attempted but falls back to AppleScript).
3. Add io.LimitReader around all JSON decoding from external sources
   (gui/updater.go:129).
4. Add a context timeout to the update download HTTP request
   (gui/updater.go:281).
5. Clamp SimilarityPct between 1 and 100 in scanner/photo.go. E2. Concurrency
   Safety
6. Replace global log.SetOutput in gui/state.go with a package-level
   \*log.Logger instance.
7. Fix the TOCTOU race in gui/cache_cleaner.go:cleanPath() by combining
   measurement and deletion into a single atomic operation.
8. Add a bounded semaphore pattern to the statPass feeder in
   diskanalyzer/walker.go. E3. Code Quality
9. Consolidate duplicate formatSize/FormatBytes into internal/fsutil.
10. Remove FindDuplicates deprecated wrapper and update the GUI to use the new
    Scanner interface.
11. Replace bubble sort in diskanalyzer/render_cli.go with sort.Slice.
12. Remove goto from cmd/root.go.
13. Clean up build constraints — remove old // +build lines. E4. Tooling

- Already present: .golangci.yml with errcheck, govet, ineffassign, staticcheck,
  unused, gofmt, goimports. This is good.
- Missing: Add gosec (security scanning) and nilness (nil pointer analysis) to
  the linter config.
- Add govulncheck to the CI pipeline to detect known CVEs in dependencies.
- Run golangci-lint with --enable-all at least once to catch additional issues
  (then dial back as needed). E5. Dependency Hygiene
- The vendored fyne/ directory (~892 Go files) is unusual. Ensure go mod vendor
  is used consistently rather than a manual copy. As of Go 1.25, vendoring
  should be managed via go mod vendor.
- golang.org/x/term is a minor dependency for a single terminal-width function.
  Evaluate if os.Stdout fallback to 80 columns would suffice (eliminating the
  dependency).
- github.com/nfnt/resize (indirect via goimagehash) is unmaintained — last
  commit 2016. Consider if this is a security risk. E6. Test Coverage
- Good test infrastructure exists. Consider adding:
- Integration tests for the GUI updater (mocked HTTP server)
- Fuzz tests for path validation in internal/trash
- Property-based tests for BK-tree correctness
