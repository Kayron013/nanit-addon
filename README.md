# Nanit Bridge — Home Assistant Add-on

Local bridge for Nanit baby monitors: the camera's video streams over your
LAN into Home Assistant, and sensors (temperature, humidity, night mode,
motion/sound) plus night-light and standby switches appear automatically via
MQTT. Authentication and camera control go through Nanit's cloud API; the
video itself never leaves your network.

## Install

1. HA → Settings → Add-ons → Add-on Store → ⋮ → Repositories → add
   `https://github.com/Kayron013/nanit-addon`.
2. Install **Nanit Bridge**.
3. Start it. If the log doesn't show your host's LAN IP on the
   `RTMP ingest advertised to camera` line, set the `rtmp_listen_ip`
   option and restart.
4. Open `http://<host-ip>:8080` (or the add-on's **Open Web UI** button),
   log in with your Nanit account, and enter the 2FA code Nanit emails you.

## Entities

With a Mosquitto broker present, MQTT discovery creates everything
automatically: a device per baby with temperature, humidity, night-mode,
stream-alive, last-motion and last-sound sensors plus night-light and
standby switches.

## Camera

Recommended path: install the [AlexxIT go2rtc](https://github.com/AlexxIT/hassio-addons)
add-on and add this stream (baby UID is shown in the bridge dashboard):

```yaml
streams:
  nursery:
    - "ffmpeg:rtmp://<host-ip>:1935/local/<baby_uid>#video=copy#audio=aac#audio=opus"
```

Keep `audio=aac` before `audio=opus` — reversing them makes casts/Google
devices silent.

Then add a **Generic Camera** integration with stream source
`rtsp://<host-ip>:8554/nursery` and still image
`http://<host-ip>:1984/api/frame.jpeg?src=nursery`.

## More

Options reference and troubleshooting: [nanit_bridge/DOCS.md](nanit_bridge/DOCS.md).
