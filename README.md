# Nanit Bridge — Home Assistant Add-on Repository

Local-first bridge for Nanit baby monitors, packaged as a Home Assistant OS add-on.

Vendored from [daleiii/nanit-web](https://github.com/daleiii/nanit-web) at commit `88f3d37` (see `nanit_bridge/UPSTREAM.md`). The source is committed into this repository so the add-on builds locally on your machine, independent of upstream repo or Docker Hub availability.

## Install

1. Push this repository to GitHub (update `url` in `repository.yaml` and `config.yaml`).
2. HA → Settings → Add-ons → Add-on Store → ⋮ → Repositories → add your repo URL.
3. Install **Nanit Bridge**, set `rtmp_listen_ip`, start, complete login at `http://<host>:8080`.
4. Add entities using `nanit_bridge/ha-nanit-package.yaml`.

## Updating from upstream

Diff upstream against `nanit_bridge/nanit-src/`, review, copy changes in, bump `version` in `config.yaml`, note it in `CHANGELOG.md`. The pinned commit lives in `nanit-src/.upstream-commit`.
