# Contributing to DupClean

Thank you for considering contributing to DupClean! Here are the guidelines to follow.

## Development Setup

### Prerequisites

- Go 1.22 or later
- Git
- Make (optional, but recommended)

### Linux Dependencies

```bash
# Ubuntu/Debian
sudo apt-get install -y libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install -y mesa-libGL-devel libX11-devel libXrandr-devel libXi-devel libXcursor-devel libXinerama-devel libXfixes-devel

# Arch Linux
sudo pacman -S --noconfirm mesa libx11 libxrandr libxi libxcursor libxinerama libxfixes
```

### macOS Dependencies

```bash
brew install glfw
```

### Clone and Build

```bash
git clone https://github.com/PopolQue/dupclean.git
cd dupclean
make build
```

## Making Changes

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

- Follow Go best practices
- Keep changes focused and atomic
- Add tests for new functionality

### 3. Run Tests and Linting

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run linter
make lint

# Format code
make fmt
```

### 4. Commit Your Changes

We follow conventional commit messages:

```bash
git commit -m "feat: add new feature description"
git commit -m "fix: fix bug description"
git commit -m "docs: update documentation"
git commit -m "refactor: refactor code description"
git commit -m "test: add tests for feature"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then open a pull request on GitHub.

## Code Style

### Go Conventions

- Run `go fmt ./...` before committing
- Run `go vet ./...` to catch common errors
- Use meaningful variable and function names
- Keep functions small and focused
- Add comments for complex logic (not for obvious code)

### Project Structure

```
dupclean/
├── main.go              # Entry point
├── gui/
│   └── app.go           # GUI implementation
├── scanner/
│   └── scanner.go       # Duplicate detection logic
├── ui/
│   └── ui.go            # Terminal UI
└── ...
```

### Testing

- Write tests for new functionality
- Aim for meaningful coverage (not 100% for the sake of it)
- Test edge cases and error conditions
- Use table-driven tests where appropriate

Example:

```go
func TestFindDuplicates(t *testing.T) {
    tests := []struct {
        name     string
        files    []string
        expected int
    }{
        {"no duplicates", []string{"a.txt", "b.txt"}, 0},
        {"one duplicate", []string{"a.txt", "a.txt"}, 1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Pull Request Guidelines

### Before Submitting

- [ ] Code is formatted (`go fmt ./...`)
- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Commit message follows conventions
- [ ] Changes are tested locally

### PR Description

Please include:

1. **What** this PR changes
2. **Why** these changes are needed
3. **How** you tested the changes
4. **Screenshots** if UI changes (optional)

### Review Process

1. CI runs automatically on your PR
2. Maintainers will review your code
3. Address any feedback
4. Once approved, your PR will be merged

## Reporting Issues

### Bug Reports

Include:

- DupClean version
- Operating system and version
- Steps to reproduce
- Expected vs actual behavior
- Screenshots/logs if applicable

### Feature Requests

Include:

- What problem you're trying to solve
- How you envision the feature working
- Any alternatives you've considered

## Questions?

- Open an issue for questions
- Check existing issues for similar topics
- Read the [README](README.md) and [documentation](docs/)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
