# OIDC Provider Configuration

KubeDeck OAuth can run in `stub` mode (default) or `oidc` mode (production-oriented).

## Mode

- `KUBEDECK_OAUTH_MODE=stub`
  - Uses internal stub provider (MVP/demo).
- `KUBEDECK_OAUTH_MODE=oidc`
  - Uses OIDC discovery + OAuth2 code exchange.
  - Powered by `go-oidc` and `golang.org/x/oauth2`.

## Required Env (OIDC mode)

- `KUBEDECK_OIDC_ISSUER` (e.g. `https://idp.example.com/realms/main`)
- `KUBEDECK_OIDC_CLIENT_ID`
- `KUBEDECK_OIDC_CLIENT_SECRET` (if required by IdP)
- `KUBEDECK_OIDC_REDIRECT_URL` (callback URL registered in IdP)

Optional:
- `KUBEDECK_OIDC_SCOPES` (comma-separated, default: `openid,profile,email`)
- `KUBEDECK_OAUTH_PROVIDER` (display/provider label, default `oidc`)

## API Flow

1. `GET /api/auth/oauth/url`
   - Returns `auth_url` and one-time `state`.
2. Browser redirects to IdP authorize URL.
3. Frontend calls `POST /api/auth/oauth/callback` with:
   - `code`
   - `state`
   - `tenant_code` (optional)
4. Backend verifies `state` and exchanges `code`.
5. Backend verifies `id_token` and creates KubeDeck session token.

## Security Notes

- `state` is one-time and expires in 10 minutes.
- Invalid or replayed `state` is rejected with `invalid_state`.
- OIDC claim mapping currently defaults roles to `viewer` if no role/group claim is present.
