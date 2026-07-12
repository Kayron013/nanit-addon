# Changelog

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
