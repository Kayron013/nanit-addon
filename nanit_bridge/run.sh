#!/usr/bin/with-contenv bashio
# ==============================================================================
# Nanit Bridge add-on: map Supervisor options -> NANIT_* environment variables
# and launch the bridge. Upstream source: daleiii/nanit-web @ 88f3d37 (vendored).
# ==============================================================================
set -e

# --- Fixed paths --------------------------------------------------------------
export NANIT_DATA_DIR="/data"
export NANIT_SESSION_FILE="/data/session.json"
export NANIT_HTTP_PORT="8080"
export NANIT_LOG_LEVEL="$(bashio::config 'log_level')"

# --- RTMP: the camera must reach us on the HOST's LAN IP + mapped host port ---
RTMP_HOST_PORT="$(bashio::addon.port '1935/tcp')"
if ! bashio::var.has_value "${RTMP_HOST_PORT}"; then
    bashio::exit.nok "RTMP port 1935/tcp is disabled in the add-on network config. Re-enable it — the Nanit camera pushes its stream to this port."
fi

RTMP_IP="$(bashio::config 'rtmp_listen_ip')"
if ! bashio::var.has_value "${RTMP_IP}"; then
    bashio::log.info "rtmp_listen_ip not set — attempting auto-detection via Supervisor network info"
    # First IPv4 address of the default interface, CIDR suffix stripped.
    RTMP_IP="$(bashio::network.ipv4_address | head -n1 | cut -d'/' -f1 || true)"
fi
if ! bashio::var.has_value "${RTMP_IP}"; then
    bashio::exit.nok "Could not determine this host's LAN IP. Set 'rtmp_listen_ip' in the add-on configuration to your Home Assistant host's IP address (the camera must be able to reach it)."
fi
export NANIT_RTMP_ENABLED="true"
export NANIT_RTMP_ADDR="${RTMP_IP}:${RTMP_HOST_PORT}"
bashio::log.info "RTMP ingest advertised to camera at ${NANIT_RTMP_ADDR}"

# --- MQTT: manual override > Supervisor-discovered Mosquitto > disabled -------
if bashio::config.true 'mqtt_enabled'; then
    if bashio::config.has_value 'mqtt_broker_url'; then
        export NANIT_MQTT_BROKER_URL="$(bashio::config 'mqtt_broker_url')"
        export NANIT_MQTT_USERNAME="$(bashio::config 'mqtt_username')"
        export NANIT_MQTT_PASSWORD="$(bashio::config 'mqtt_password')"
        bashio::log.info "MQTT: using manually configured broker ${NANIT_MQTT_BROKER_URL}"
    elif bashio::services.available 'mqtt'; then
        MQTT_HOST="$(bashio::services 'mqtt' 'host')"
        MQTT_PORT="$(bashio::services 'mqtt' 'port')"
        export NANIT_MQTT_BROKER_URL="tcp://${MQTT_HOST}:${MQTT_PORT}"
        export NANIT_MQTT_USERNAME="$(bashio::services 'mqtt' 'username')"
        export NANIT_MQTT_PASSWORD="$(bashio::services 'mqtt' 'password')"
        bashio::log.info "MQTT: auto-discovered broker at ${NANIT_MQTT_BROKER_URL}"
    else
        bashio::log.warning "MQTT enabled but no broker configured and none discoverable via the Supervisor. Continuing without MQTT — set 'mqtt_broker_url' or install the Mosquitto add-on."
    fi
    if bashio::var.has_value "${NANIT_MQTT_BROKER_URL:-}"; then
        export NANIT_MQTT_ENABLED="true"
        export NANIT_MQTT_PREFIX="$(bashio::config 'mqtt_topic_prefix')"
        export NANIT_MQTT_CLIENT_ID="$(bashio::config 'mqtt_topic_prefix')"
    else
        export NANIT_MQTT_ENABLED="false"
    fi
else
    export NANIT_MQTT_ENABLED="false"
fi

# --- Events polling ------------------------------------------------------------
if bashio::config.true 'events_polling'; then
    export NANIT_EVENTS_POLLING="true"
    export NANIT_EVENTS_POLLING_INTERVAL="$(bashio::config 'events_polling_interval')"
else
    export NANIT_EVENTS_POLLING="false"
fi

# --- History -------------------------------------------------------------------
if bashio::config.true 'history_enabled'; then
    export NANIT_HISTORY_ENABLED="true"
    export NANIT_HISTORY_RETENTION_DAYS="$(bashio::config 'history_retention_days')"
else
    export NANIT_HISTORY_ENABLED="false"
fi

bashio::log.info "Starting Nanit Bridge (dashboard on port 8080; complete Nanit login + 2FA there on first run)"
exec /app/bin/nanit
