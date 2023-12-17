package web

import (
	"dev.hackerman.me/artheon/artheon-rpc/models"
	"github.com/google/uuid"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1 << 20
)

type WebsocketMessageType int32

const (
	PushMessageType     WebsocketMessageType = 1 << iota // Simple one-way messages.
	RequestMessageType                                   // RPC requests. Awaits the corresponding response to be returned from other side. Optionally includes RPC args.
	ResponseMessageType                                  // RPC responses. Should be returned for the corresponding request optionally holding the payload.
)

type WebsocketTopic int32

const (
	SystemTopic    WebsocketTopic = 1 << iota // System messages (connection, presence).
	ChatTopic                                 // Chat related messages.
	AnalyticsTopic                            // Analytics related messages.
	VivoxTopic                                // Vivox related messages.
)

// Text chat categories
const (
	CategorySystem  string = "system"
	CategoryGeneral string = "general"
	CategorySpace   string = "space"
	CategoryServer  string = "server"
	CategoryPrivate string = "private"
	CategoryUnknown string = "unknown"
)

const (
	MessageNotifyUserJoinedChannel string = "userJoinedChannel"
	MessageNotifyUserLeftChannel   string = "userLeftChannel"
)

const (
	PresenceStatusPlaying   string = "playing"
	PresenceStatusAvailable string = "available"
	PresenceStatusAway      string = "away"
	PresenceStatusOffline   string = "offline"
)

const (
	ConnectMethod            string = "connect"            // Connect to the server. Initial websocket connection handshake.
	PresenceUpdateMethod     string = "presenceUpdate"     // Connect to the server. Initial websocket connection handshake.
	ChannelSubscribeMethod   string = "channelSubscribe"   // Subscribe to existing channel. Used to connect to known channel, e.g. global or space channels.
	ChannelUnsubscribeMethod string = "channelUnsubscribe" // Unsubscribe from the channel. Used when user leaves space to stop to receive local space messages.
	ChannelSendMethod        string = "channelSend"        // Send the message to the channel.
	UserChangeNameMethod     string = "userChangeName"     // Change user's name.
	UserActionMethod         string = "userAction"         // Report user action.
	VivoxGetLoginTokenMethod string = "vivoxGetLoginToken" // Request vivox token.
	VivoxGetJoinTokenMethod  string = "vivoxGetJoinToken"  // Request vivox token.
	VivoxMuteMethod          string = "vivoxMute"          // Request vivox server-to-server action.
	VivoxUnmuteMethod        string = "vivoxUnmute"        // Request vivox server-to-server action.
	VivoxKickMethod          string = "vivoxKick"          // Request vivox server-to-server action.
)

type WebsocketPayload struct {
	Status    string       `json:"status,omitempty"`
	Message   string       `json:"message,omitempty"`
	Sender    *models.User `json:"sender,omitempty"`
	ChannelId string       `json:"channelId,omitempty"`
	Category  string       `json:"category,omitempty"`
}

type WebsocketMessage struct {
	Id      uuid.UUID            `json:"id,omitempty"`      // Used to match RPC requests and their corresponding responses.
	Type    WebsocketMessageType `json:"type,omitempty"`    // Used to determine the type of the message (simple or RPC).
	Topic   WebsocketTopic       `json:"topic,omitempty"`   // Used to determine the purpose of the message.
	Method  string               `json:"method,omitempty"`  // Used to determine the handler method to process.
	Payload interface{}          `json:"payload,omitempty"` // Used for responses and push messages.
	Args    interface{}          `json:"args,omitempty"`    // Used for requests.
}
