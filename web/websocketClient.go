package web

import (
	"bytes"
	"dev.hackerman.me/artheon/artheon-rpc/models"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// An instance created for each websocket connection.
type WebsocketClient struct {
	// Unique ID of the socket client.
	Id uuid.UUID
	// The websocket server.
	server *WebsocketServer
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbounds messages.
	send chan []byte
	// Requests
	requests map[uuid.UUID]time.Time
	// Channels the client is subscribed for
	channels []uuid.UUID
	// Rpc request handlers
	handlers map[string]websocketRequestHandler
	// Owning user
	user *models.User
	// Is disconnecting
	disconnecting chan bool
}

type websocketRequestHandler func(client *WebsocketClient, message *WebsocketMessage, topic WebsocketTopic, method string, args interface{}) error

// Register an RPC handler function.
func (client *WebsocketClient) registerHandler(topic WebsocketTopic, method string, handler websocketRequestHandler) {
	name := fmt.Sprintf("%d.%s", topic, method)
	client.handlers[name] = handler
}

// Subscribe the client for the channel.
func (client *WebsocketClient) addChannelSubscription(channelId uuid.UUID) {
	for _, c := range client.channels {
		if c == channelId {
			log.Printf("client {%s} had already subscribed for the channel {%s}", client.Id.String(), channelId.String())
			return
		}
	}

	client.channels = append(client.channels, channelId)
	log.Printf("client {%s} subscribed for the channel {%s} of category {%s}", client.Id.String(), channelId.String(), getCategoryByChannelId(&channelId))
}

// Unsubscribe the client from the channel.
func (client *WebsocketClient) removeChannelSubscription(channelId uuid.UUID) {
	for idx, v := range client.channels {
		if v == channelId {
			client.channels = append(client.channels[0:idx], client.channels[idx+1:]...)
			log.Printf("client {%s} unsubscribed from the channel {%s}", client.Id.String(), channelId.String())
			return
		}
	}
	log.Printf("client {%s} was not subscribed for the channel {%s}", client.Id.String(), channelId.String())
}

// Close the websocket connection.
func (client *WebsocketClient) closeWebsocket(reason string) {
	_ = client.conn.WriteMessage(websocket.CloseMessage, []byte(reason))

	client.disconnecting <- true

	// Unsubscribe the client from all channels
	for _, v := range client.channels {
		client.removeChannelSubscription(v)
	}

	close(client.send)
}

// Connection established callback. Called after client connection. Sends special message to the frontend.
func (client *WebsocketClient) onConnectionEstablished() {
	// Send a request for connection.
	//if err := client.SendPushMessage(SystemTopic, WebsocketPayload{Message: "Welcome to Artheon WebSockets RPC server. Please authenticate to proceed."}); err != nil {
	//      log.Errorf("got an error sending a message to the websocket: %s", err.Error())
	//      return
	//}

	log.Printf("connection established to client %s", client.Id)
}

// Reads messages from the websocket connection to the server.
// Runs in a per-connection goroutine. The application ensures that there is at most one
// reader per connection by executing all reads from this goroutine.
func (client *WebsocketClient) goSocketRead() {
	defer func() {
		client.server.unregister <- client
	}()

	client.conn.SetReadLimit(maxMessageSize)

	_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(message string) error {
		_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()

		if err != nil {
			log.Errorf("got an unexpected websocket close error: %s", err.Error())
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		websocketMessage, err := client.server.serializer.deserialize(message)

		if err != nil {
			log.Errorf("got an invalid websocket message: %s", err.Error())
		}

		err = client.onMessageReceived(websocketMessage)

		if err != nil {
			log.Errorf("got an error reading data from socket: %s", err)
		}
	}
}

// Writes messages from the server to the websocket connection.
// Runs in a per-connection goroutine. The application ensures that there is at most one
// writer per connection by executing all writes from this goroutine.
func (client *WebsocketClient) goSocketWrite() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				log.Printf("unable to get the message from the client send channel: %s", message)
				// The server closed the channel.
				_ = client.conn.Close()
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Errorf("got an error trying to get websocket connection writer: %s", err.Error())
				return
			}
			_, _ = w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				_, _ = w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				log.Errorf("got an error trying to close websocket connection writer: %s", err.Error())
				return
			}
		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Errorf("got an error trying to write a ping message to a websocket: %s", err.Error())
				return
			}
		}
	}
}

func (client *WebsocketClient) goRequestDeadline() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = client.conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			for message, expiresAt := range client.requests {
				if time.Now().Sub(expiresAt) > 0 {
					log.Errorf("got an timeout processing request: %+v", message)
					delete(client.requests, message)
				}
			}
		case disconnecting := <-client.disconnecting:
			if disconnecting {
				return
			}
		}
	}
}
