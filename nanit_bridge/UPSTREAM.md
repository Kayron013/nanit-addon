# Upstream provenance

- Source: https://github.com/daleiii/nanit-web
- Pinned commit: 88f3d37 (sole commit on main as of 2026-07-11)
- License: MIT (see nanit-src lineage: adam.stanek/nanit -> indiefan/home_assistant_nanit -> daleiii/nanit-web)
- Vendored: cmd/, pkg/, frontend/, scripts/, go.mod, go.sum
- Local modifications: NONE. All add-on behavior is packaging-layer only
  (Dockerfile, run.sh, config.yaml).

## Source audit notes (2026-07-11)
- Outbound endpoints: api.nanit.com only (auth, token refresh, babies,
  messages). No telemetry, no third-party endpoints.
- MQTT: plain state topics <prefix>/babies/<uid>/<key> and command topics
  .../night_light/switch, .../standby/switch (payload "true"/"false").
  NO Home Assistant MQTT discovery, contrary to upstream README.
- Health endpoint /api/health/<baby_uid> requires a UID (400 without) —
  hence TCP watchdog rather than HTTP.
- Frontend: Next.js static export, root-absolute asset/API paths —
  not ingress-compatible without patching.
