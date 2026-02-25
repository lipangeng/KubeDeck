# Development Notes

## 2026-02-25

- Some historical commits in this branch contain shell-escaped message rendering artifacts (backtick segments stripped in commit body display).
- Functional code changes are preserved and covered by subsequent commits and regression tests.
- Follow-up commits now use file-based commit messages to avoid shell interpolation issues.
