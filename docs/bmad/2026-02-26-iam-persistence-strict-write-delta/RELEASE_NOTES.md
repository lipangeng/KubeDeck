# Release Notes Delta

## Backend
- IAM write endpoints now fail with `500 persistence_failed` when persistence cannot initialize/write.
- Added regression tests for group/member/invite create failure paths.
