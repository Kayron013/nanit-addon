# Upstream provenance

- Source: https://github.com/daleiii/nanit-web
- Pinned commit: 88f3d37 (sole commit on main as of 2026-07-11)
- License: MIT (see nanit-src lineage: adam.stanek/nanit -> indiefan/home_assistant_nanit -> daleiii/nanit-web)
- Vendored: cmd/, pkg/, frontend/, scripts/, go.mod, go.sum

## Local patches (policy changed 2026-07-12)

The vendored source is no longer byte-identical to upstream. Patches are
kept minimal, one commit each (`git log -- nanit_bridge/nanit-src` is the
authoritative list, diffable against the pinned upstream commit):

1. **MQTT discovery** (`pkg/mqtt/discovery.go` + hooks in `pkg/mqtt/mqtt.go`,
   `pkg/mqtt/opts.go`, `cmd/nanit/main.go`, `pkg/app/app.go`): publish
   retained HA discovery configs per baby (upstream README claimed this but
   never implemented it). Gated by env `NANIT_MQTT_DISCOVERY` (default true).
2. **Log redaction** (`pkg/app/api_handlers.go`): the web login/2FA handlers
   logged the plain-text password, the MFA code, and full Nanit API response
   bodies (access/refresh tokens, email, phone). Removed.
3. **Live baby list** (`pkg/app/serve_react.go`): web API routes captured the
   baby list at server startup, so a first-run login showed "no babies" until
   the process was restarted. Handlers now read the session store per-request.

## Source audit notes (2026-07-11)
- Outbound endpoints: api.nanit.com only (auth, token refresh, babies,
  messages). No telemetry, no third-party endpoints.
- MQTT: plain state topics <prefix>/babies/<uid>/<key> and command topics
  .../night_light/switch, .../standby/switch (payload "true"/"false").
  Upstream had NO Home Assistant MQTT discovery, contrary to its README —
  added as local patch 1 (see above) in add-on 1.1.0.
- Health endpoint /api/health/<baby_uid> requires a UID (400 without) —
  hence TCP watchdog rather than HTTP.
- Frontend: Next.js static export, root-absolute asset/API paths —
  not ingress-compatible without patching.
