package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/indiefan/home_assistant_nanit/pkg/baby"
	"github.com/rs/zerolog/log"
)

// discoveryPrefix - topic root Home Assistant watches for MQTT discovery
// config messages (HA's default; configurable on the HA side only).
const discoveryPrefix = "homeassistant"

// BabyResolver - looks up baby metadata by UID so discovery devices can be
// named after the baby instead of the raw UID.
type BabyResolver func(babyUID string) *baby.Baby

// SetBabyResolver - registers the resolver used when building discovery
// device info. Optional; without it devices are named by UID.
func (conn *Connection) SetBabyResolver(resolver BabyResolver) {
	conn.babyResolver = resolver
}

// publishDiscoveryConfigs - announces one baby's entities to Home Assistant
// as retained MQTT discovery configs, so the entities appear without manual
// YAML. The RTMP camera feed cannot be described by MQTT discovery and still
// needs a Generic Camera / ffmpeg entity on the HA side.
func (conn *Connection) publishDiscoveryConfigs(babyUID string) {
	deviceName := fmt.Sprintf("Nanit %v", babyUID)
	if conn.babyResolver != nil {
		if b := conn.babyResolver(babyUID); b != nil && b.Name != "" {
			deviceName = fmt.Sprintf("Nanit %v", b.Name)
		}
	}

	device := map[string]interface{}{
		"identifiers":  []string{fmt.Sprintf("nanit_%v", babyUID)},
		"name":         deviceName,
		"manufacturer": "Nanit",
	}

	stateTopic := func(key string) string {
		return fmt.Sprintf("%v/babies/%v/%v", conn.Opts.TopicPrefix, babyUID, key)
	}

	// State payloads are fmt.Sprintf("%v") of the Go values: booleans are
	// "true"/"false", timestamps are integer seconds since epoch.
	timestampTemplate := "{{ (value | int | as_datetime).isoformat() }}"

	entities := []struct {
		component string
		object    string
		config    map[string]interface{}
	}{
		{"sensor", "temperature", map[string]interface{}{
			"name":                "Temperature",
			"state_topic":         stateTopic("temperature"),
			"device_class":        "temperature",
			"unit_of_measurement": "°C",
			"state_class":         "measurement",
		}},
		{"sensor", "humidity", map[string]interface{}{
			"name":                "Humidity",
			"state_topic":         stateTopic("humidity"),
			"device_class":        "humidity",
			"unit_of_measurement": "%",
			"state_class":         "measurement",
		}},
		{"sensor", "motion_timestamp", map[string]interface{}{
			"name":           "Last motion",
			"state_topic":    stateTopic("motion_timestamp"),
			"device_class":   "timestamp",
			"value_template": timestampTemplate,
			"icon":           "mdi:motion-sensor",
		}},
		{"sensor", "sound_timestamp", map[string]interface{}{
			"name":           "Last sound",
			"state_topic":    stateTopic("sound_timestamp"),
			"device_class":   "timestamp",
			"value_template": timestampTemplate,
			"icon":           "mdi:ear-hearing",
		}},
		{"binary_sensor", "is_night", map[string]interface{}{
			"name":        "Night mode",
			"state_topic": stateTopic("is_night"),
			"payload_on":  "true",
			"payload_off": "false",
			"icon":        "mdi:weather-night",
		}},
		{"binary_sensor", "is_stream_alive", map[string]interface{}{
			"name":         "Stream alive",
			"state_topic":  stateTopic("is_stream_alive"),
			"payload_on":   "true",
			"payload_off":  "false",
			"device_class": "connectivity",
		}},
		{"switch", "night_light", map[string]interface{}{
			"name":          "Night light",
			"state_topic":   stateTopic("night_light"),
			"command_topic": stateTopic("night_light/switch"),
			"payload_on":    "true",
			"payload_off":   "false",
			"icon":          "mdi:lightbulb-night",
		}},
		{"switch", "standby", map[string]interface{}{
			"name":          "Standby",
			"state_topic":   stateTopic("standby"),
			"command_topic": stateTopic("standby/switch"),
			"payload_on":    "true",
			"payload_off":   "false",
			"icon":          "mdi:video-off",
		}},
	}

	// Every entity — sensors included — shares the camera's availability:
	// sensor data rides the same websocket as commands, so a dead camera
	// link means stale sensors too. "Stale but green" is the failure mode
	// availability exists to eliminate. availability_mode "all" also takes
	// everything unavailable when the bridge itself is down (LWT).
	availability := []map[string]interface{}{
		{"topic": conn.bridgeStatusTopic()},
		{"topic": conn.babyStatusTopic(babyUID)},
	}

	for _, entity := range entities {
		entity.config["unique_id"] = fmt.Sprintf("nanit_%v_%v", babyUID, entity.object)
		entity.config["device"] = device
		entity.config["availability"] = availability
		entity.config["availability_mode"] = "all"

		topic := fmt.Sprintf("%v/%v/nanit_%v/%v/config", discoveryPrefix, entity.component, babyUID, entity.object)
		payload, err := json.Marshal(entity.config)
		if err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Unable to marshal discovery config")
			continue
		}

		token := conn.client.Publish(topic, 0, true, payload)
		if token.Wait(); token.Error() != nil {
			log.Error().Err(token.Error()).Str("topic", topic).Msg("Unable to publish discovery config")
		}
	}

	log.Info().Str("baby_uid", babyUID).Str("device", deviceName).Msg("Published MQTT discovery configs")
}

// subscribeToHomeAssistantStatus - listens for Home Assistant's birth
// message and republishes availability plus a full state snapshot for every
// known baby. State topics are deliberately non-retained, so after an HA
// restart entities would otherwise sit at `unknown` until the next organic
// state change.
func (conn *Connection) subscribeToHomeAssistantStatus() {
	topic := discoveryPrefix + "/status"

	handler := func(_ MQTT.Client, msg MQTT.Message) {
		if string(msg.Payload()) != availabilityOnline {
			return
		}

		log.Info().Msg("Home Assistant came online — republishing availability and state")

		go func() {
			// HA's birth message can precede its MQTT subscriptions being
			// fully active; give it a moment before replaying state.
			time.Sleep(2 * time.Second)

			for _, babyUID := range conn.StateManager.GetKnownBabyUIDs() {
				state := conn.StateManager.GetBabyState(babyUID)
				conn.publishAvailability(conn.babyStatusTopic(babyUID), state.GetIsWebsocketAlive())
				conn.publishBabyState(babyUID, *state)
			}
		}()
	}

	if token := conn.client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Str("topic", topic).Msg("Failed to subscribe to Home Assistant status topic")
	}
}
