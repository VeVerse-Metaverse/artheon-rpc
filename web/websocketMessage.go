package web

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

func (client *WebsocketClient) onMessageReceived(websocketMessage *WebsocketMessage) (err error) {

	log.Printf("received message, client: {%s}, message: {%s}, type: {%d}, , topic: {%d}, method: {%s}, payload: {%v}, args: {%v}", client.Id, websocketMessage.Id, websocketMessage.Type, websocketMessage.Topic, websocketMessage.Method, websocketMessage.Payload, websocketMessage.Args)

	if websocketMessage == nil {
		return errors.New("received empty pointer instead of websocket message")
	}

	switch websocketMessage.Type {

	case PushMessageType:

		// Server does not react to push messages.
		log.Printf("ignoring push message")

	case RequestMessageType:

		log.Printf("processing request message")

		topic := websocketMessage.Topic
		method := websocketMessage.Method
		handlerName := fmt.Sprintf("%d.%s", topic, method)

		handler, ok := client.handlers[handlerName]
		if !ok {
			return errors.New(fmt.Sprintf("handler not found for the websocket message method: %s", method))
		}

		args := websocketMessage.Args

		// Call the assigned request handler function with specified args.
		err := handler(client, websocketMessage, topic, method, args)

		if err != nil {
			log.Errorf("got an error sending rpc response to socket: %s", err.Error())
			return err
		}

	case ResponseMessageType:
		log.Printf("processing response message")
		delete(client.requests, websocketMessage.Id)
	}

	return nil
}

func (client *WebsocketClient) sendRequestMessage(topic WebsocketTopic, method string, args interface{}) (err error) {

	request := WebsocketMessage{
		Id:     uuid.New(),
		Type:   RequestMessageType,
		Topic:  topic,
		Method: method,
		Args:   args,
	}

	// Save request for timeout detection.
	client.requests[request.Id] = time.Now().Add(pongWait)

	encodedMessage, err := client.server.serializer.serialize(&request)

	if err != nil {
		log.Errorf("error serializing an rpc request message: %s", err.Error())
		return err
	}

	log.Printf("sendRequestMessage %s, %s", request.Method, request.Id.String())

	client.send <- encodedMessage

	return
}

func (client *WebsocketClient) sendResponseMessage(websocketMessage *WebsocketMessage, payload interface{}) (err error) {
	response := WebsocketMessage{
		Id:      websocketMessage.Id, // The request ID should match the response.
		Type:    ResponseMessageType,
		Topic:   websocketMessage.Topic,
		Method:  websocketMessage.Method,
		Payload: payload,
	}

	encodedMessage, err := client.server.serializer.serialize(&response)

	if err != nil {
		log.Errorf("error serializing an rpc response message: %s", err.Error())
		return err
	}

	log.Printf("sendResponseMessage %s, %s", response.Method, response.Id.String())

	client.send <- encodedMessage

	return
}

// Sends the text message to the frontend via websocket.
func (client *WebsocketClient) SendPushMessage(topic WebsocketTopic, payload interface{}) (err error) {

	message := WebsocketMessage{
		Id:      uuid.New(),
		Type:    PushMessageType,
		Topic:   topic,
		Payload: payload,
	}

	serializedMessage, err := client.server.serializer.serialize(&message)

	log.Printf("SendPushMessage %s, %s", message.Topic, message.Id.String())

	client.send <- serializedMessage

	return
}
