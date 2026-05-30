# How DupClean Works

## Audio & Byte Mode (4-Stage Algorithm)

1. **Size Pre-Filter** — Groups files by size (instant, skips 99% of non-duplicates)
2. **Partial Hash** — Hashes first 8KB of potential matches (very fast)
3. **Full SHA-256 Hash** — Hashes entire file content for exact matches
4. **Byte Comparison** — Final verification to guarantee 100% accuracy

### Performance

Up to **100x faster** than naive hashing because

- Files with unique sizes are never hashed
- Files with different content at the start are rejected after 8KB
- Only likely duplicates undergo full hashing and verification

## Photo Mode (Perceptual Hashing)

1. **Decode Image** — Load and normalize the image
2. **Perceptual Hash** — Compute a 64-bit fingerprint based on image structure
3. **Hamming Distance** — Compare hashes to find similar images
4. **Group by Similarity** — Cluster images above similarity threshold

### What it catches

| | Yes | No |
| - | - | -- |
| Resized images | x | |
| Re-encoded at different quality | x | |
| Slight color adjustments | x | |
| Cropped versions | x | |
| Heavily edited or composite images | | x |
