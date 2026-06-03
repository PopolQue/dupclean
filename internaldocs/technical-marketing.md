# Plan: Technical Marketing Strategy for "GitHub Stardom"

## Objective

Increase DupClean's visibility within the developer and creator communities to
reach a target of 2,000+ stars by highlighting its unique engineering rigor and
"Swiss Army Knife" utility.

## Strategy: "Show, Don't Just Tell"

### 1. Visual Overhaul (The "First 3 Seconds" Rule)

- **High-Quality GIFs:** Create three optimized GIFs for the top of the README:
  - **The Disk Analyzer:** A 5-second loop of the treemap rendering.
  - **Photo Similarity:** A split-screen showing two slightly different photos
    being matched.
  - **Smart Select:** The Duplicate Finder resolving 10 groups in one click.
- **Feature Icons:** Replace the plain table in the README with a "Feature Grid"
  using high-resolution icons or emojis to make it scannable.

### 2. "Engineering Deep Dive" Blog Posts

Write and cross-post (Dev.to, Hashnode, Medium) technical articles that appeal
to the Go community:

- **Article 1:** "How I reached 90% Test Coverage on a Cross-Platform File
  Deletion Tool." (Focus on the mocking framework and safety).
- **Article 2:** "Perceptual Hashing in Go: Finding Similar Photos at Scale."
  (Focus on the `scanner/photo.go` implementation).
- **Article 3:** "Building Native GUIs in Go: Why I chose Fyne over Electron."
  (Focus on performance and binary size).

### 3. Community Outreach (Targeted Channels)

- **Reddit:**
  - `r/golang`: Post a "Showcase" thread focusing on the concurrency of the Disk
    Analyzer.
  - `r/selfhosted`: Position DupClean as the "Privacy-First" alternative to
    cloud-based cleanup tools.
  - `r/musicproduction`: Highlight the Audio Mode and `afplay` integration.
- **Hacker News:** Submit the "Engineering Deep Dive" on test coverage; HN loves
  stories about "Safety-Critical" software.
- **Fyne Showcase:** Formally submit DupClean to the
  [Fyne Apps](https://apps.fyne.io/) gallery.

### 4. "Star-Aware" GitHub Optimization

- **Social Preview:** Design a "Social Preview" image for the repo settings
  (1280x640) that shows the UI and the "90% Coverage" badge.
- **Interactive Demo:** Use "GitHub Pages" or a "WebAssembly" build of the GUI
  (Fyne supports WASM) to let users "try" the UI in their browser without
  installing.
- **Product Hunt Launch:** Schedule a "Product Hunt" launch once the "Community
  Recipes" system is live to trigger a second wave of growth.
