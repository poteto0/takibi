---
name: golang-tdd
description: Use this skill for Go source code implementation tasks — regardless of what language the user writes in. Trigger when the user wants to: add a method or attribute to an existing type, implement a new feature or behavior, fix a bug in source code, or write/expand test cases for existing code. This skill applies to any direct modification of .go files. Do not trigger for: documentation or README edits, dependency version bumps, git config changes (.gitignore etc.), PR review, or conceptual Go language questions.
---

# Golang TDD

Every code change starts with a failing test. The cycle is always: **Red → Green → Refactor**.

## TDD Cycle

Repeat this loop until the entire spec is satisfied:

### 1. RED — Write a failing test

Write the next test case that describes one unit of desired behavior. Run it and confirm it fails.

```bash
go test ./... -run TestYourNewTest -v
```

The test must fail before you write implementation. A test that passes without code isn't a TDD test.

### 2. GREEN — Write minimum code to pass, then refactor

Write only enough code to make the test pass — no more.

```bash
go test ./... -run TestYourNewTest -v
```

Once green, refactor freely: clean up duplication, improve naming, extract helpers in test or source code. The tests stay green throughout.

```bash
go test ./...  # confirm nothing broke after refactor
```

### 3. Back to RED — Next test case

Add the next test case for the next piece of behavior. Loop back to step 1. Keep looping until the spec is fully implemented.

---

## Go Test Patterns

### Table-driven tests (preferred for multiple cases)

```go
func TestParser_Parse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "foo", "FOO", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### HTTP handler tests (use `net/http/httptest`)

```go
func TestHandler_Get(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/hello", nil)
    w := httptest.NewRecorder()

    handler(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Async behavior (use `assert.Eventually`)

```go
assert.Eventually(t, func() bool {
    return triggered
}, 1*time.Second, 10*time.Millisecond)
```

### Test naming convention

```
TestTypeName_MethodName        — top-level
t.Run("descriptive scenario")  — subtests
```

White-box tests (same package as the code) for internal access.
Black-box tests (`package foo_test`) for public API contracts.

---

## Workflow Checklist

When asked to implement anything, follow this sequence:

1. **Understand the interface** — read the relevant types and interfaces before writing anything
2. **Loop until the spec is satisfied:**
   - Write the next test case (RED) — run it, confirm it fails
   - Write minimum code to make it pass (GREEN) — run it, confirm it passes
   - Refactor test or source code as needed — run `go test ./...` to confirm still green
   - Move to the next test case

---

## Common Pitfalls

- **Don't write implementation before the test.** If you catch yourself writing both at once, stop — write the test file first, run it to see red, then write the implementation.
- **Don't test implementation details.** Test behavior through the public/package interface, not unexported helpers.
- **Prefer real instances over mocks** for internal code. Mocks that diverge from real behavior let tests pass while production breaks.
