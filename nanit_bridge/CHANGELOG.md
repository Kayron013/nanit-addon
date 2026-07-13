# Changelog

## 1.1.1
- Add "Open Web UI" button on the add-on page (webui declaration)

## 1.1.0

First release with local patches to the vendored source (policy change —
see UPSTREAM.md for the patch list; previously the source was byte-identical
to upstream daleiii/nanit-web @ 88f3d37).

- **MQTT auto-discovery**: entities (temperature, humidity, night mode,
  stream alive, last motion/sound, night light + standby switches) now
  appear in Home Assistant automatically as a "Nanit <baby name>" device.
  `ha-nanit-package.yaml` is no longer needed except for the camera entity
  (or use the Generic Camera UI integration). Disable with the new
  `mqtt_discovery` option. If you previously installed the manual package,
  remove it before updating — duplicate unique_ids conflict.
- **Security: redacted log output.** The web login flow logged the
  plain-text Nanit password, the MFA code, and full API response bodies
  (access/refresh tokens, email, phone number) at info level. If you ever
  shared old add-on logs, treat those tokens as exposed (changing your
  Nanit password rotates them).
- **Fixed first-run "no babies" in the dashboard**: the web API served the
  baby list captured at startup, so a fresh login required an add-on restart
  before the dashboard worked. It now reads the live session.
- Validated: `go build ./...`, `go vet`, and upstream tests pass on the
  modified source (also the first-ever compile validation of this codebase).

## 1.0.2
- Fix "Unable to access the API, forbidden" at startup: run.sh shebang is now
  `#!/usr/bin/with-contenv bashio` so the s6-overlay base image passes
  SUPERVISOR_TOKEN (and the rest of the container environment) to the script
- Add `hassio_api: true` so bashio::network.ipv4_address can query
  /network/info for RTMP listen-IP auto-detection (endpoint is role-gated,
  unlike the /addons/self/* calls)

## 1.0.1
- Fix Supervisor build failure: declare BUILD_FROM before the first FROM so
  the final stage's `FROM ${BUILD_FROM}` can resolve it (an ARG between
  stages is scoped to the preceding stage and invisible to later FROM lines)

## 1.0.0
- Initial add-on packaging of daleiii/nanit-web @ 88f3d37 (vendored source, local build)
- Supervisor options mapped to NANIT_* env; Mosquitto auto-discovery via mqtt:want
- RTMP listen IP auto-detection with manual override
- TCP watchdog on dashboard port; session persisted in /data (backed up)
- Included manual MQTT entity package template (upstream has no MQTT discovery)
