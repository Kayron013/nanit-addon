# Changelog

## 1.2.2

- **Stop streaming retry loops from accumulating across reconnects.**
  Observed in production (2026-08-12) against an unresponsive camera: the
  request rate climbed steadily — 12 requests in the first 40-second
  connection window, 69 by the ninth, over 200/minute after ten minutes —
  and reset only on add-on restart. `requestLocalStreaming` retries a
  timed-out start request forever, and its only exit was the per-baby
  `IsWebsocketAlive` flag. That flag tracks the *newest* connection, so
  after the liveness monitor recycled a wedged socket it read true again
  within a second and stale loops never noticed their own connection had
  died. Every reconnect left another loop behind, each still sending on a
  dead socket. Connections now carry an explicit closed marker, set when
  the handler tears down, and a retry loop stops as soon as the connection
  it was spawned for is retired. Interleaved, out-of-order request IDs in
  the logs were the tell — each stale connection kept its own counter.
- **Send teardown streaming commands once.** PAUSED/STOPPED requests are
  best-effort signals issued while a connection is closing; they were
  retried on the same infinite loop as a start request, adding traffic
  nothing was waiting on. They now get a single attempt.
- **Release pending response handlers on timeout.** `resHandlers` entries
  were only removed when a matching response arrived, so against a camera
  that answers nothing every request left a permanent entry — unbounded
  map growth for the life of the connection, worst exactly when the retry
  loops were at their most active.

## 1.2.1

- **Detect the one-way websocket wedge.** Observed in production (2026-07-17):
  the camera keeps pushing sensor data — so inbound looks healthy and the
  1.2.0 silence probe never fires — while every request (streaming,
  night light, standby) times out unanswered, producing endless
  "Streaming request timeout, trying again" log loops and unresponsive
  toggles. The liveness monitor now also probes after 3 consecutive
  request timeouts; a failed probe forces the same reconnect path. Any
  matched response (even a late one) resets the counter, so a healthy
  connection is never reconnect-looped.

## 1.2.0

MQTT/websocket robustness release — fixes the "night light toggle dead
until add-on restart" failure and makes entity availability honest.

- **Websocket liveness with probe-before-kill.** Half-dead camera
  connections (commands blackholed, no disconnect event ever fired) are
  now detected: >60s of inbound silence triggers a lightweight
  GET_SENSOR_DATA probe; a failed/timed-out probe (15s) declares the
  link dead and forces a reconnect. Silence alone never kills a healthy
  quiet connection, and the probe round-trips through Nanit's cloud to
  the camera, so it also catches a dead cloud↔camera segment.
- **Availability topics.** Bridge-level `<prefix>/bridge/status`
  (retained, with MQTT Last Will for ungraceful deaths) and per-camera
  `<prefix>/babies/<uid>/status` (retained, driven by websocket
  liveness). All discovered entities reference both with
  `availability_mode: all` — entities now show *unavailable* in HA when
  the bridge is down or the camera link is dead, instead of stale
  values and toggles that lie. Sensors share the camera availability
  deliberately: sensor data rides the same websocket as commands.
- **State republish on HA restart.** The bridge listens for Home
  Assistant's MQTT birth message and replays availability plus a full
  state snapshot for every known baby (~2s after HA comes up). No more
  `unknown` entities until the next organic state change.
- **Distinct MQTT client ID.** The bridge previously connected with
  client ID equal to the topic prefix (`nanit`), inviting broker
  session-stealing collisions; it now defaults to `nanit-bridge`
  (config: `NANIT_MQTT_CLIENT_ID`). Note: because the client uses
  persistent sessions, the old `nanit` broker session is abandoned —
  harmless, the broker expires it.

## 1.1.2
- Add add-on icon

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
