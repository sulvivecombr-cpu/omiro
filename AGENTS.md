# Notes
- Go/Echo app on :8080 (hardcoded), mapped to host 3000. Needs Redis (compose service). No secrets required.
- Live reload via `air` (installed at container start; first boot takes a minute).
- Session token secret is hardcoded in middleware/session_token.go.
