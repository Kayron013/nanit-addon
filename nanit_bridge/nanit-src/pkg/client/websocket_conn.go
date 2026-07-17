package client

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sacOO7/gowebsocket"
	"github.com/indiefan/home_assistant_nanit/pkg/utils"
	"google.golang.org/protobuf/proto"
)

// WebsocketMessageHandler - message handler
type WebsocketMessageHandler func(*Message, *WebsocketConnection)

// WebsocketConnection - ready websocket connection
type WebsocketConnection struct {
	socket *gowebsocket.Socket

	msgHandlersMu sync.RWMutex
	msgHandlers   []WebsocketMessageHandler

	resHandlersMu sync.RWMutex
	resHandlers   map[int32]unhandledRequest

	lastRequestID int32

	// lastInbound - unix nanos of the most recent inbound message of any
	// type; the liveness monitor uses this to detect half-dead connections
	// that the transport layer never reports (see websocket.go)
	lastInbound int64

	// consecutiveTimeouts - requests that timed out with no response since
	// the last answered request. Detects the one-way wedge the silence
	// window cannot: the camera keeps pushing sensor data (inbound looks
	// healthy) while every request goes unanswered — observed in production
	// as endless "Streaming request timeout" loops and dead night-light
	// commands. Reset by any matched response, even a late one: a late
	// response still proves the request channel works.
	consecutiveTimeouts int32
}

// NewWebsocketConnection - constructor
func NewWebsocketConnection(socket *gowebsocket.Socket) *WebsocketConnection {
	return &WebsocketConnection{
		socket:        socket,
		resHandlers:   make(map[int32]unhandledRequest),
		lastRequestID: 0,
		lastInbound:   time.Now().UnixNano(),
	}
}

// LastInbound - time of the most recent inbound message
func (conn *WebsocketConnection) LastInbound() time.Time {
	return time.Unix(0, atomic.LoadInt64(&conn.lastInbound))
}

// ConsecutiveTimeouts - requests that timed out since the last response
func (conn *WebsocketConnection) ConsecutiveTimeouts() int32 {
	return atomic.LoadInt32(&conn.consecutiveTimeouts)
}

// RegisterMessageHandler - registers handler which will be called whenever new message is received
func (conn *WebsocketConnection) RegisterMessageHandler(handler WebsocketMessageHandler) {
	conn.msgHandlersMu.Lock()
	conn.msgHandlers = append(conn.msgHandlers, handler)
	conn.msgHandlersMu.Unlock()
}

// SendMessage - low-level helper for sending raw message
// Note: Use SendRequest() for requests
func (conn *WebsocketConnection) SendMessage(m *Message) error {
	var msg *zerolog.Event

	if *m.Type == Message_KEEPALIVE {
		msg = log.Trace()
	} else {
		msg = log.Debug()
	}

	msg.Stringer("data", m).Msg("Sending message")

	bytes, err := getMessageBytes(m)
	if err != nil {
		return fmt.Errorf("failed to marshal websocket message: %w", err)
	}
	log.Trace().Bytes("rawdata", bytes).Msg("Sending data")

	conn.socket.SendBinary(bytes)
	return nil
}

// SendRequest - sends request to the cam and returns await function. Await function waits for the response and returns it
func (conn *WebsocketConnection) SendRequest(reqType RequestType, requestData *Request) func(time.Duration) (*Response, error) {
	// Build request
	id := atomic.AddInt32(&conn.lastRequestID, 1)

	requestData.Id = utils.ConstRefInt32(id)
	requestData.Type = RequestType(reqType).Enum()

	m := &Message{
		Type:    Message_Type(Message_REQUEST).Enum(),
		Request: requestData,
	}

	// Response handling
	resC := make(chan *Response, 1)

	conn.resHandlersMu.Lock()
	conn.resHandlers[id] = unhandledRequest{
		Request: m.Request,
		HandleResponse: func(res *Response) {
			select {
			case <-resC:
				return // Channel already closed (ie. timeout)
			default:
				resC <- res
			}
		},
	}
	conn.resHandlersMu.Unlock()

	// Send request
	if err := conn.SendMessage(m); err != nil {
		log.Error().Err(err).Msg("Failed to send websocket message")
		// Return an awaiter that immediately returns the error
		return func(timeout time.Duration) (*Response, error) {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
	}

	// Return awaiter
	return func(timeout time.Duration) (*Response, error) {
		timer := time.NewTimer(timeout)

		select {
		case <-timer.C:
			close(resC)
			atomic.AddInt32(&conn.consecutiveTimeouts, 1)
			return nil, errors.New("Request timeout")
		case res := <-resC:
			close(resC)
			timer.Stop()

			if res.StatusCode == nil {
				return res, errors.New("No status code received")
			} else if *res.StatusCode != 200 {
				if res.GetStatusMessage() != "" {
					return res, errors.New(res.GetStatusMessage())
				}

				return res, fmt.Errorf("Unexpected status code %v", *res.StatusCode)
			}

			return res, nil
		}
	}
}

type unhandledRequest struct {
	Request        *Request
	HandleResponse func(response *Response)
}

func (conn *WebsocketConnection) handleResponse(r *Response) {
	requestID := *r.RequestId
	requestType := *r.RequestType

	conn.resHandlersMu.RLock()
	unhandledReqCandidate, ok := conn.resHandlers[requestID]
	conn.resHandlersMu.RUnlock()

	if ok && requestType == *unhandledReqCandidate.Request.Type {
		conn.resHandlersMu.Lock()
		delete(conn.resHandlers, requestID)
		conn.resHandlersMu.Unlock()

		atomic.StoreInt32(&conn.consecutiveTimeouts, 0)
		unhandledReqCandidate.HandleResponse(r)
	}
}

func (conn *WebsocketConnection) handleMessage(m *Message) {
	atomic.StoreInt64(&conn.lastInbound, time.Now().UnixNano())

	if *m.Type == Message_RESPONSE && m.Response != nil {
		conn.handleResponse(m.Response)
	}

	conn.msgHandlersMu.RLock()
	subscribedHandlers := make([]WebsocketMessageHandler, len(conn.msgHandlers))
	copy(subscribedHandlers, conn.msgHandlers)
	conn.msgHandlersMu.RUnlock()

	for _, handler := range subscribedHandlers {
		handler(m, conn)
	}
}

func getMessageBytes(data *Message) ([]byte, error) {
	out, err := proto.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Unable to marshal data")
		return nil, err
	}

	return out, nil
}
