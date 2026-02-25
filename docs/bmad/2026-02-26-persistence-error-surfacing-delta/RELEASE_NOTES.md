# Release Notes Delta

- Rebuild API now fails fast with 500 when persistence cannot initialize/write.
- Added strict persistence helper layer for error propagation.
- Added regression coverage for persistence failure path.
