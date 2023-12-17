package web

import (
	"database/sql"
	"fmt"
	"os"

	"dev.hackerman.me/artheon/artheon-rpc/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const SystemChannelId string = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
const GlobalChannelId string = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"

// Websocket server instance.
var WebsocketServerInstance *WebsocketServer

// Start websocket server.
func init() {
	WebsocketServerInstance = newWebsocketServer()
	go WebsocketServerInstance.start()
}

type PrivateChannelInfo struct {
	// Id of the user who initiated the private channel
	Host uuid.UUID

	// Id of the user invited to the private channel
	Guest uuid.UUID
}

type ChannelInfo struct {
	// Global channel
	SystemChannel uuid.UUID

	// Global channel
	GeneralChannel uuid.UUID

	// Registered space channels.
	SpaceChannels []uuid.UUID

	// Registered server channels.
	ServerChannels []uuid.UUID

	// Registered user channels, allows up to 2 users to join for private conversation.
	PrivateChannels map[uuid.UUID]PrivateChannelInfo
}

// A single instance at the server side.
type WebsocketServer struct {
	ChannelInfo

	// Known users
	Users map[uuid.UUID]models.User

	// Known spaces
	Spaces map[uuid.UUID]models.Space

	// Database connection
	Db *sql.DB

	// Registered clients.
	Clients map[uuid.UUID]*WebsocketClient

	// Channel to broadcast system messages to all clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *WebsocketClient

	// Unregister requests from the clients.
	unregister chan *WebsocketClient

	// Message serializer
	serializer *WebsocketMessageSerializer
}

func newWebsocketServer() *WebsocketServer {
	server := &WebsocketServer{
		ChannelInfo: ChannelInfo{
			SystemChannel:   uuid.MustParse(SystemChannelId),
			GeneralChannel:  uuid.MustParse(GlobalChannelId),
			SpaceChannels:   make([]uuid.UUID, 0),
			PrivateChannels: make(map[uuid.UUID]PrivateChannelInfo),
		},
		Clients:    make(map[uuid.UUID]*WebsocketClient),
		broadcast:  make(chan []byte),
		register:   make(chan *WebsocketClient),
		unregister: make(chan *WebsocketClient),
		serializer: newWebsocketMessageSerializer(),
	}

	return server
}

func (server *WebsocketServer) registerClient(client *WebsocketClient) {
	server.Clients[client.Id] = client
}

func (server *WebsocketServer) unregisterClient(client *WebsocketClient) {
	if _, ok := server.Clients[client.Id]; ok {
		//client.closeWebsocket("unregister client")
		delete(server.Clients, client.Id)
	}
}

func (server *WebsocketServer) broadcastMessage(message []byte) {
	// Send a message to each client.
	for _, client := range server.Clients {
		client.send <- message
	}
}

func (server *WebsocketServer) start() {

	var err error

	db_user := os.Getenv("DB_USER")
	if db_user == "" {
		db_user = "postgres"
	}
	db_pass := os.Getenv("DB_PASS")
	if db_pass == "" {
		db_pass = "postgres"
	}
	db_host := os.Getenv("DB_HOST")
	if db_host == "" {
		db_host = "127.0.0.1"
	}
	db_port := os.Getenv("DB_PORT")
	if db_port == "" {
		db_port = "5432"
	}
	db_name := os.Getenv("DB_NAME")
	if db_name == "" {
		db_name = "veverse"
	}
	db_url := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", db_user, db_pass, db_host, db_port, db_name)

	if server.Db, err = sql.Open("postgres", db_url); err != nil {
		log.Fatal(err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer func() {
		server.Db.Close()
	}()

	for {
		select {
		// Register a client.
		case client := <-server.register:
			server.registerClient(client)

		// Unregister a client.
		case client := <-server.unregister:
			server.unregisterClient(client)

		// On message.
		case message := <-server.broadcast:
			server.broadcastMessage(message)
		}
	}
}
