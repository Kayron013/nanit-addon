package mqtt

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

const (
	availabilityOnline  = "online"
	availabilityOffline = "offline"
)

// bridgeStatusTopic - bridge-level availability. Retained; also registered
// as the MQTT Last Will so the broker flips it to offline if the bridge
// dies without a clean disconnect.
func (conn *Connection) bridgeStatusTopic() string {
	return fmt.Sprintf("%v/bridge/status", conn.Opts.TopicPrefix)
}

// babyStatusTopic - per-camera availability, driven by websocket liveness.
func (conn *Connection) babyStatusTopic(babyUID string) string {
	return fmt.Sprintf("%v/babies/%v/status", conn.Opts.TopicPrefix, babyUID)
}

// publishAvailability - publishes a retained online/offline payload
func (conn *Connection) publishAvailability(topic string, online bool) {
	payload := availabilityOffline
	if online {
		payload = availabilityOnline
	}

	token := conn.client.Publish(topic, 1, true, payload)
	if token.Wait(); token.Error() != nil {
		log.Error().Err(token.Error()).Str("topic", topic).Msg("Unable to publish availability")
	}
}
