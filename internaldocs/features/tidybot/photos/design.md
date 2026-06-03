# TidyBot: Photo Organization

**Status:** In-Development
**Author:** Gemini CLI
**Last Updated:** 2026-05-30

## 1. Overview & Scope

Automated photo sorting using EXIF/GPS metadata via extended scanner functionality.

## 2. Technical Architecture

- **Memory Management:** EXIF parsing via streaming (do not load full image to heap).
- **Concurrency Model:** Independent processing of image EXIF blocks.
- **Syscall/IO Strategy:** Buffered reads for EXIF header parsing.

## 3. Interface Definition (API Contract)

```go
type EXIFData struct {
    DateTaken time.Time
    GPS       Location
}

func ExtractEXIF(path string) (*EXIFData, error)
```

## 4. Error Handling & Safety

- **Propagation:** Graceful failure if EXIF is missing; log as "Uncategorized".
- **Safety Audits:** GPS data sanitized for privacy.

## 5. Performance Constraints

- **Latency:** Extraction < 100ms per file.

## 6. Verification Plan

- Unit tests with varied EXIF profiles (Canon/Nikon/iPhone).

---
[Link back to Master Overview](./../design.md)
