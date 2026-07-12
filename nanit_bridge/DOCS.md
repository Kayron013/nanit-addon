# Nanit Bridge

Runs a local bridge for Nanit baby monitors: it authenticates against the Nanit cloud API, commands the camera to push its RTMP stream to this add-on on your LAN, restreams it for Home Assistant/ffmpeg consumption, and publishes sensor state (temperature, humidity, day/night) plus night-light and standby controls over MQTT.

Cloud dependency note: authentication and camera control go through `api.nanit.com` (verified: this is the only outbound endpoint in the source). The video stream itself travels camera → this add-on entirely on your LAN.

## First-time setup

1. Set `rtmp_listen_ip` to your Home Assistant host's LAN IP (e.g. `192.168.1.50`). Leave empty to attempt auto-detection — check the log to confirm what was detected. This must be an address the camera can reach; never `127.0.0.1`.
2. Start the add-on.
3. Open `http://<host-ip>:8080`, enter your Nanit email/password, then the 2FA code emailed to you. The session token is stored in `/data/session.json` (persisted, included in HA backups).
4. Note your `baby_uid` from the dashboard or the add-on log.

## Home Assistant entities

Since add-on 1.1.0 the bridge publishes **MQTT discovery** configs (local patch; upstream never implemented it despite its README): once a baby's first state update arrives, a "Nanit <name>" device appears in Settings → Devices & Services → MQTT with temperature, humidity, night mode, stream-alive, last motion/sound, and night-light/standby switches. No YAML needed. Disable with the `mqtt_discovery` option if you prefer manual entities (`ha-nanit-package.yaml` remains as a template for that case — don't use both, the duplicate `unique_id`s will conflict).

The **camera** is the one entity discovery can't create. Add it in the UI: Settings → Devices & Services → Add Integration → **Generic Camera**, stream source `http://<host-ip>:8080/api/stream/hls/<baby_uid>/playlist.m3u8` (unauthenticated HLS, served by the bridge). Do **not** point Generic Camera at the RTMP URL — HA Core's stream worker (PyAV) is unstable with RTMP sources and can crash Home Assistant outright. The RTMP URL (`rtmp://<host-ip>:1935/local/<baby_uid>`) remains fine for external consumers like VLC or go2rtc.

If the Mosquitto add-on is installed, broker credentials are wired up automatically; a manual `mqtt_broker_url` (e.g. `tcp://host:1883`) overrides discovery.

## Options

| Option | Description |
|---|---|
| `rtmp_listen_ip` | Host LAN IP advertised to the camera for RTMP push. Empty = auto-detect. |
| `log_level` | trace / debug / info / warn / error |
| `mqtt_enabled` | Publish state + subscribe to command topics |
| `mqtt_broker_url` | Manual broker override, e.g. `tcp://192.168.1.50:1883` |
| `mqtt_username` / `mqtt_password` | Manual broker credentials |
| `mqtt_topic_prefix` | Topic prefix (default `nanit`) |
| `mqtt_discovery` | Auto-create HA entities via MQTT discovery (default `true`) |
| `events_polling` / `events_polling_interval` | Poll Nanit cloud for event messages |
| `history_enabled` / `history_retention_days` | Local SQLite history for the dashboard |

## Known limitations

- **No ingress**: the dashboard frontend uses root-absolute paths and cannot be proxied under HA ingress without patching. Access it directly at port 8080 on your LAN. Do not port-forward it.
- **Camera connection limits**: Nanit cameras cap concurrent local streaming connections. Video may fail while sensors keep working; the bridge retries automatically. This is Nanit-side throttling, not an add-on fault.
- **Reverse-engineered protocol**: Nanit can break this at any time by changing their API (they did once before, with mandatory 2FA).

## Security

The refresh token in `/data/session.json` grants full access to your Nanit account. Keep this add-on LAN-only. Optional dashboard password protection can be enabled in the web UI; reset it with `docker exec` against the add-on container if locked out.
