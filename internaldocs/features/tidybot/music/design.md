# TidyBot: Music Organization

**Status:** In-Development
**Author:** Gemini CLI
**Last Updated:** 2026-05-30

## 1. Overview & Scope

Specialized support for organizing music collections by leveraging ID3/FLAC metadata tags.

## 2. Technical Architecture

- **Memory Management:** Metadata struct reuse to prevent excessive allocations during batch processing.
- **Concurrency Model:** Stateless processing (metadata extraction) allows concurrent evaluation of multiple files.
- **Syscall/IO Strategy:** Read-only access to audio files for metadata extraction using low-overhead parsers.

## 3. Interface Definition (API Contract)

```go
type MusicMetadata struct {
    Artist, Album, Year, Label string
}

func ExtractMusicTags(path string) (*MusicMetadata, error)
```

## 4. Error Handling & Safety

- **Propagation:** Fallback to generic organization if metadata is corrupt/missing.
- **Safety Audits:** Sanitization of metadata strings for filesystem path usage (prevent path traversal).

## 5. Performance Constraints

- **Latency:** Extraction < 50ms per file.

## 6. Verification Plan

- Unit tests against varied audio formats/metadata corruption.

---
[Link back to Master Overview](./../design.md)
