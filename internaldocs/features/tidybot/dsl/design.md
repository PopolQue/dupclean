# TidyBot: DSL Parser & Tokenizer

**Status:** In-Development **Author:** Gemini CLI **Last Updated:** 2026-05-30

## 1. Overview

High-performance parser for the TidyBot DSL, transforming human-readable rules
into an Abstract Syntax Tree (AST) suitable for the
[Rule Engine](./../engine/design.md).

## 2. Formal Grammar (BNF)

```bnf
<rule>        ::= "WHEN" <condition_list> "THEN" <action_list>
<condition>   ::= <field> <operator> <value>
<field>       ::= "ext" | "name" | "size" | "age" | "dir" | "tag" | "metadata(" <string> ")"
<action>      ::= <action_type> <args>
```

## 3. API Contract (Go)

```go
type ASTNode interface { Evaluate(fs.FileInfo) bool }

type Rule struct {
    Conditions []Condition
    Actions    []Action
}

type Parser interface {
    Parse(dsl string) (*Rule, error)
}
```

## 4. Technical Specifications

- **Tokenizer:** Lexical analysis using `text/scanner` to handle comments and
  whitespace efficiently.
- **Error Handling:** Returns `ErrSyntax` with line/column information on
  failure.
- **Performance:** AST caching enabled for frequently evaluated rules to
  minimize parsing overhead.

## 5. Testing Strategy

- Positive/Negative grammar tests (100+ cases).
- Benchmarking parsing throughput vs. regex-based alternatives.

---

[Link back to Master Overview](./../design.md)
