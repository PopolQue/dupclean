# Master Prompt

You are a senior Go (Golang) engineer and systems architect performing a full repository audit.

Analyze the entire Go codebase in this repository. Your focus is on production-grade quality, correctness, security, and maintainability.

---

## 1. Code Quality & Go Best Practices

Evaluate the codebase against idiomatic Go standards:

* Correct use of Go project structure (cmd/, internal/, pkg/, etc.)
* Proper separation of concerns and package boundaries
* Idiomatic Go patterns (interfaces, composition over inheritance)
* Error handling quality (no ignored errors, proper wrapping with fmt.Errorf / errors.Join)
* Context usage correctness (context propagation, cancellation handling)
* Concurrency correctness (goroutines, race conditions, channel usage, sync primitives)
* Avoidance of global state and hidden dependencies
* Function and package design (small, focused functions; cohesive packages)
* Code duplication and unnecessary abstraction
* Naming conventions (Go style guidelines)
* Dependency hygiene (go.mod cleanliness, unnecessary dependencies, vendor issues)

Flag anti-patterns such as:

* “interface pollution” (defining interfaces too early or unnecessarily)
* over-abstraction
* excessive layering without benefit
* misuse of pointers vs values

---

## 2. Security (Go-specific focus)

Identify security vulnerabilities and unsafe patterns:

* Input validation issues (especially in HTTP handlers / APIs)
* Injection risks (SQL, command execution, template injection)
* Unsafe use of `os/exec`
* Insecure HTTP handling (missing timeouts, unsafe clients, no TLS config checks)
* Authentication/authorization flaws
* Improper secret management (env vars, config files, logs)
* SSRF risks in HTTP clients
* Path traversal vulnerabilities
* Insecure deserialization (JSON, gob, etc.)
* Dependency vulnerabilities in Go modules
* Missing rate limiting / abuse protection in APIs

Pay special attention to:

* net/http usage correctness
* database/sql usage safety
* third-party HTTP clients
* any crypto usage (ensure modern, correct primitives)

---

## 3. Performance & Concurrency

Assess Go-specific runtime and performance issues:

* Goroutine leaks
* Improper channel usage (deadlocks, blocking issues)
* Unbounded concurrency
* Inefficient allocations or excessive copying
* Unnecessary locking or contention (sync.Mutex misuse)
* Poor HTTP server/client configuration (timeouts, keep-alives)
* Missing batching or inefficient I/O patterns

---

## 4. UX / API Design (if applicable)

If the repo includes APIs or frontend-facing logic:

* API consistency and ergonomics
* Error response structure consistency
* Status code correctness
* Pagination/filtering design
* Developer experience of API usage
* Clarity and predictability of endpoints

---

## Output Format

### A. Executive Summary

* Overall codebase health
* Key architectural strengths and weaknesses
* Maturity level (junior / mid / senior / production-grade Go service)

---

### B. Critical Issues (Must Fix)

* Security vulnerabilities
* Concurrency bugs or race conditions
* Data loss risks or crash scenarios
* Broken API or correctness issues

---

### C. Important Improvements

* Structural or architectural improvements
* Go idiom violations affecting maintainability
* Performance issues
* Dependency or module problems

---

### D. Minor / Style Issues

* Naming inconsistencies
* Small refactors
* Lint-level improvements (go fmt / golint / staticcheck type issues)

---

### E. Concrete Recommendations

* Specific refactoring suggestions (mention packages/files when possible)
* Suggested Go tooling improvements (e.g., `staticcheck`, `golangci-lint`, `govulncheck`)
* Concurrency safety improvements
* Security hardening steps

---

## Constraints

* Analyze the entire repository, not a subset
* Be concrete and reference actual packages/files when possible
* Prioritize real production impact over theoretical issues
* Prefer Go idioms over generic software engineering advice
* If something cannot be determined from the code, explicitly state uncertainty
* Avoid vague statements — every finding must be actionable
