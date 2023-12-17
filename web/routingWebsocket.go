package web

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1 << 20,
	WriteBufferSize: 1 << 20,
}

func (s webServer) initWebsocketRoutes() {
	// Websocket routing.
	r := s.router.PathPrefix("/ws").Subrouter()

	r.Path("").
		Methods("GET").
		HandlerFunc(handleWebsocket).
		Name("ws")
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Errorf("got an error trying to upgrade to websocket protocol: %s", err.Error())
		return
	}

	// Create and register a client.
	client := &WebsocketClient{Id: uuid.New(),
		server:   WebsocketServerInstance,
		conn:     conn,
		send:     make(chan []byte, 256),
		handlers: make(map[string]websocketRequestHandler),
		requests: make(map[uuid.UUID]time.Time),
	}

	client.registerHandler(SystemTopic, ConnectMethod, connectHandler)
	client.registerHandler(SystemTopic, PresenceUpdateMethod, presenceUpdateHandler)
	client.registerHandler(SystemTopic, UserChangeNameMethod, userChangeNameHandler)

	client.registerHandler(ChatTopic, ChannelSendMethod, channelMessageHandler)
	client.registerHandler(ChatTopic, ChannelSubscribeMethod, channelSubscribeHandler)
	client.registerHandler(ChatTopic, ChannelUnsubscribeMethod, channelUnsubscribeHandler)

	client.registerHandler(AnalyticsTopic, UserActionMethod, userActionHandler)

	client.registerHandler(VivoxTopic, VivoxGetLoginTokenMethod, vivoxHandler)
	client.registerHandler(VivoxTopic, VivoxGetJoinTokenMethod, vivoxHandler)
	client.registerHandler(VivoxTopic, VivoxMuteMethod, vivoxHandler)
	client.registerHandler(VivoxTopic, VivoxUnmuteMethod, vivoxHandler)
	client.registerHandler(VivoxTopic, VivoxKickMethod, vivoxHandler)

	client.server.register <- client

	go client.goSocketWrite()
	go client.goSocketRead()
	go client.goRequestDeadline()

	client.onConnectionEstablished()
}
