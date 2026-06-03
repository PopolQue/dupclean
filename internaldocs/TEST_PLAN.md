# Test Plan for DupClean

## Current Coverage Status

| Package         | Coverage | Status                    |
| --------------- | -------- | ------------------------- |
| scanner         | 91.3%    | ✅ Excellent              |
| diskanalyzer    | 43.1%    | ⚠️ Needs improvement      |
| ui              | 85.6%    | ✅ Good                   |
| gui             | 32.7%    | ⚠️ Needs improvement      |
| cleaner         | 91.5%    | ✅ Excellent              |
| internal/fsutil | 80.8%    | ✅ Good                   |

## Priority 1: Cleaner Package (Critical)

The cleaner package is now well-tested.

### Tests to Add:

- [x] `cleaner/scanner_test.go` (Covered by integration/scanner tests)
- [x] `cleaner/deleter_test.go`
- [x] `cleaner/targets_test.go` (Covered by developer/system tests)

## Priority 2: GUI Package (Critical)

The GUI has complex logic but still needs more test coverage.

### Tests to Add:

- [ ] `gui/cache_cleaner_test.go`
- [ ] `gui/sidebar_test.go`
- [ ] `gui/duplicate_finder_test.go`

## Priority 3: Internal/fsutil

- [x] `internal/fsutil/measure_test.go`

## Priority 4: Improve Existing Coverage

### Scanner Package (91.3%)

- [x] Test ByteScanner.Scan()
- [x] Test PhotoScanner.Scan()
- [x] Test hashFilePartial() edge cases
- [x] Test hashFileFull() with large files
- [x] Test filesIdentical() with different scenarios

### DiskAnalyzer Package (43.1% → 80%)

- [ ] Test error handling in Walk()
- [ ] Test with symlinks
- [ ] Test ExportJSON() output format
- [ ] Test RenderCLI() output formatting

### UI Package (85.6%)

- [x] Test CLI input parsing
- [x] Test invalid user input handling
- [x] Test moveToTrash() integration

## Test Infrastructure Improvements

- [ ] Add test helpers/mocks for:
  - File system operations
  - OS-specific functions (trash, permissions)
  - Time-based operations
- [ ] Add integration tests:
  - Full scan → review → delete workflow
  - Cache cleaner workflow
  - Disk analyzer workflow

- [ ] Add benchmark tests:
  - Benchmark large directory scans
  - Benchmark hashing performance
  - Benchmark UI rendering with many items

