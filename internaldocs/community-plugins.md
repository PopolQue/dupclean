# Plan: Community "Clean Recipes" Plugin System

## Objective

Enable users to contribute and share custom cache-cleaning definitions (recipes)
without modifying the core Go source code. This transforms DupClean from a
static tool into a community-driven cleaning platform.

## Proposed Solution: YAML-Based Recipe Engine

Instead of hardcoding paths in `cleaner/target.go`, we will implement a system
that loads cleaning definitions from external YAML files.

### 1. Recipe Schema Design

Define a standard format for a "Recipe":

```yaml
name: 'Ableton Live Cache'
category: 'Developer'
description: 'Clears temporary project renders and index files'
platform: 'darwin' # or "windows", "linux", "all"
targets:
  - path: '~/Library/Application Support/Ableton/Live */Caches'
    type: 'directory'
    recursive: true
  - path: '~/Music/Ableton/Undo'
    type: 'directory'
    safe_to_delete: false # Requires extra confirmation
```

### 2. Implementation Steps

- **Core Engine:** Create a `recipe_loader.go` in the `cleaner/` package to
  parse YAML files using `gopkg.in/yaml.v3`.
- **Path Resolution:** Implement a robust glob-pattern resolver that handles
  environment variables (e.g., `%AppData%` on Windows, `~` on macOS) and
  wildcards.
- **Plugin Directory:**
  - **Local:** Load from `~/.config/dupclean/recipes/*.yaml`.
  - **Remote:** Implement a `dupclean recipes update` command that fetches the
    latest community-vetted recipes from a central GitHub repository (e.g.,
    `dupclean/community-recipes`).
- **GUI Integration:** Update the Cache Cleaner tab to dynamically generate
  checkboxes based on loaded recipes.

### 3. Verification & Testing

- **Unit Tests:** Test the YAML parser with valid/invalid schemas.
- **Safety Tests:** Ensure glob patterns cannot resolve to system-critical paths
  (e.g., `/etc`, `C:\Windows`).
- **Integration Test:** Mock a filesystem and verify the engine correctly
  identifies files based on a sample recipe.

### 4. Community Rollout

- Create a `RECEPIS.md` guide explaining how to write and submit a recipe.
- Add a "Submit a Recipe" button in the GUI that links to a GitHub Issue
  template.
