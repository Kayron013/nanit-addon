# Changelog

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
