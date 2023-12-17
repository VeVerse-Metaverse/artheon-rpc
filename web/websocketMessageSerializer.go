package web

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type WebsocketMessageSerializer struct {
}

func newWebsocketMessageSerializer() *WebsocketMessageSerializer {
	return &WebsocketMessageSerializer{}
}

func (serializer *WebsocketMessageSerializer) serialize(websocketMessage *WebsocketMessage) ([]byte, error) {

	encodedString, err := json.Marshal(websocketMessage)

	if err != nil {
		log.Errorf("got an error serializing the websocket message: %s", err)
		return nil, err
	}

	return encodedString, nil
}

func (serializer *WebsocketMessageSerializer) deserialize(message []byte) (websocketMessage *WebsocketMessage, err error) {
	err = json.Unmarshal(message, &websocketMessage)

	if err != nil {
		log.Errorf("got an error unmarshalling a json encoded string: %s", err)
		return
	}

	return
}
