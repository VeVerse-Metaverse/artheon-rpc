package web

import (
	"dev.hackerman.me/artheon/artheon-rpc/models"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	handlerArgsChannel      = "channelId"
	handlerArgsUserId       = "key"
	handlerArgsMessage      = "message"
	handlerArgsPresence     = "presence"
	handlerArgsVivoxPayload = "vivoxPayload"
	handlerStatusOk         = "ok"
	handlerStatusError      = "error"
)

func connectHandler(client *WebsocketClient, websocketMessage *WebsocketMessage, _ WebsocketTopic, _ string, args interface{}) (err error) {

	//region parse args
	m, ok := args.(map[string]interface{})
	if !ok {
		err := fmt.Sprintf("unable to parse args: %+v", args)
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion parse args

	//region validate user
	userId, err := getUserId(m)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion validate user

	//region authenticate user
	err = registerSender(client, userId)

	if err != nil {
		err := fmt.Sprintf("client is not authorized, %s", err.Error())
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion authenticate user

	//region response
	result := WebsocketPayload{
		Status: handlerStatusOk,
		Sender: client.user,
	}
	err = client.sendResponseMessage(websocketMessage, result)
	//endregion response

	return err
}

func userChangeNameHandler(client *WebsocketClient, websocketMessage *WebsocketMessage, _ WebsocketTopic, _ string, args interface{}) (err error) {
	//region parse args
	m, ok := args.(map[string]interface{})
	if !ok {
		err := fmt.Sprintf("unable to parse args: %+v", args)
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion parse args

	//region validate user
	userId, err := getUserId(m)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion validate user

	//region authenticate user
	err = registerSender(client, userId)

	if err != nil {
		err := fmt.Sprintf("client is not authorized, %s", err.Error())
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion authenticate user

	//region response
	result := WebsocketPayload{
		Status: handlerStatusOk,
		Sender: client.user,
	}
	err = client.sendResponseMessage(websocketMessage, result)
	//endregion response

	return err
}

func presenceUpdateHandler(client *WebsocketClient, websocketMessage *WebsocketMessage, _ WebsocketTopic, _ string, args interface{}) (err error) {

	//region parse args
	m, ok := args.(map[string]interface{})
	if !ok {
		err := fmt.Sprintf("unable to parse args: %+v", args)
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion parse args

	//region validate user
	if client == nil {
		err := fmt.Errorf("invalid client")
		log.Errorf("invalid client")
		return err
	}

	if client.user == nil {
		err := fmt.Sprintf("client not authenticated")
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion validate user

	//region decode presence JSON
	jsonBody, err := json.Marshal(m[handlerArgsPresence])
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	var presence models.Presence

	err = json.Unmarshal(jsonBody, &presence)
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion decode presence JSON

	//region validate status
	status := presence.Status
	if !(status == PresenceStatusOffline || status == PresenceStatusAway || status == PresenceStatusAvailable || status == PresenceStatusPlaying) {
		err := fmt.Sprintf("unknown status, %s", status)
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion validate status

	//region update user presence
	err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, presence.Status, presence.SpaceId, presence.ServerId)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion update user presence

	//region response
	result := WebsocketPayload{
		Status: handlerStatusOk,
		Sender: client.user,
	}
	err = client.sendResponseMessage(websocketMessage, result)
	//endregion response

	return err
}

// Handles received chat message requests.
func channelMessageHandler(client *WebsocketClient, websocketMessage *WebsocketMessage, _ WebsocketTopic, _ string, args interface{}) (err error) {

	//region parse args
	m, ok := args.(map[string]interface{})
	if !ok {
		err := fmt.Sprintf("unable to parse args: %+v", args)
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion parse args

	//region validate user
	userId, err := getUserId(m)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	if client.user == nil || client.user.Id != userId {
		err := fmt.Sprintf("client is not authorized")
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion authenticate user

	//region validate channel subscription
	channelId, err := getChannelId(m)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	if bSubscribed := containsUUID(client.channels, *channelId); !bSubscribed {
		err := fmt.Sprintf("client tries to send to channel it is not subscribed to: client: %s, channelId: %s", client.Id, channelId)
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion validate channel subscription

	//region validate message
	if msg := m[handlerArgsMessage]; msg == nil {
		err := fmt.Sprintf("client tries to send an empty message")
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion validate message

	//region store message
	var channelName = ""

	server, err := models.GetServerById(WebsocketServerInstance.Db, *channelId)
	if err == nil {
		channelName = fmt.Sprintf("%s:%d", server.Host, server.Port)
	} else {
		space, err := models.GetSpaceById(WebsocketServerInstance.Db, *channelId)
		if err == nil {
			channelName = space.Name
		}
	}

	_ = models.AddChatMessage(WebsocketServerInstance.Db, models.ChatMessage{
		UserId:          client.user.Id,
		Message:         m[handlerArgsMessage].(string),
		ChannelId:       channelId.String(),
		ChannelName:     channelName,
		ChannelCategory: getCategoryByChannelId(channelId),
	})
	//endregion

	//region response
	payload := WebsocketPayload{
		Status:    handlerStatusOk,
		Message:   m[handlerArgsMessage].(string),
		Sender:    client.user,
		ChannelId: channelId.String(),
		Category:  getCategoryByChannelId(channelId),
	}

	result := WebsocketPayload{Status: handlerStatusOk}
	err = client.sendResponseMessage(websocketMessage, result)
	//endregion response

	//region broadcast
	broadcastMessageToChannel(*channelId, payload)
	//endregion broadcast

	return err
}

func channelSubscribeHandler(client *WebsocketClient, websocketMessage *WebsocketMessage, _ WebsocketTopic, _ string, args interface{}) (err error) {

	//region parse args
	m, ok := args.(map[string]interface{})
	if !ok {
		err := fmt.Sprintf("unable to parse args: %+v", args)
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion parse args

	//region validate user
	userId, err := getUserId(m)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	if client.user == nil || client.user.Id != userId {
		err := fmt.Sprintf("client is not authorized")
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}
	//endregion validate user

	// Get the channel the message is sent to.
	channelId, err := getChannelId(m)
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	//region subscribe to the system channel
	if WebsocketServerInstance.SystemChannel == *channelId {
		client.addChannelSubscription(*channelId)

		result := WebsocketPayload{
			Status:    handlerStatusOk,
			ChannelId: channelId.String(),
			Category:  CategorySystem,
		}
		err = client.sendResponseMessage(websocketMessage, result)

		notifyUserJoinedChannel(*channelId, client.user)

		err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, PresenceStatusAvailable, uuid.Nil, uuid.Nil)
		if err != nil {
			log.Errorf(err.Error())
			result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
			return client.sendResponseMessage(websocketMessage, result)
		}

		err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)

		return err
	}
	//endregion subscribe to the system channel

	//region subscribe to the global channel
	if WebsocketServerInstance.GeneralChannel == *channelId {
		client.addChannelSubscription(*channelId)

		result := WebsocketPayload{
			Status:    handlerStatusOk,
			ChannelId: channelId.String(),
			Category:  CategoryGeneral,
		}
		err = client.sendResponseMessage(websocketMessage, result)

		notifyUserJoinedChannel(*channelId, client.user)

		return err
	}
	//endregion subscribe to the global channel

	//region subscribe to a cached space channel
	for _, c := range WebsocketServerInstance.SpaceChannels {
		if c == *channelId {
			client.addChannelSubscription(*channelId)

			result := WebsocketPayload{
				Status:    handlerStatusOk,
				ChannelId: channelId.String(),
				Category:  CategorySpace,
			}
			err = client.sendResponseMessage(websocketMessage, result)

			notifyUserJoinedChannel(*channelId, client.user)

			err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, PresenceStatusAvailable, *channelId, client.user.Presence.ServerId)
			if err != nil {
				log.Errorf(err.Error())
				result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
				return client.sendResponseMessage(websocketMessage, result)
			}

			err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)

			return err
		}
	}
	//endregion subscribe to a cached space channel

	//region subscribe to a cached server channel
	for _, c := range WebsocketServerInstance.ServerChannels {
		if c == *channelId {
			client.addChannelSubscription(*channelId)

			result := WebsocketPayload{
				Status:    handlerStatusOk,
				ChannelId: channelId.String(),
				Category:  CategoryServer,
			}
			err = client.sendResponseMessage(websocketMessage, result)

			notifyUserJoinedChannel(*channelId, client.user)

			err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, PresenceStatusAvailable, client.user.Presence.SpaceId, *channelId)
			if err != nil {
				log.Errorf(err.Error())
				result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
				return client.sendResponseMessage(websocketMessage, result)
			}

			err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)

			return err
		}
	}
	//endregion subscribe to a cached server channel

	//region subscribe to a private channel
	for _, otherClient := range WebsocketServerInstance.Clients {

		if otherClient.user == nil {
			continue
		}

		if otherClient.user.Id == *channelId {

			if otherClient.user.Id == client.user.Id {
				err = fmt.Errorf("can not subscribe user to self, %s", otherClient.user.Id)
				log.Errorf(err.Error())
				result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
				return client.sendResponseMessage(websocketMessage, result)
			}

			var foundChannelId = findExistingPrivateChannelForUsers(client.user.Id, otherClient.user.Id)
			var newChannelId uuid.UUID

			if foundChannelId == nil {
				newChannelId, err = uuid.NewUUID()
			} else {
				newChannelId = *foundChannelId
			}

			if err != nil {
				log.Errorf(err.Error())
				result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
				return client.sendResponseMessage(websocketMessage, result)
			}

			// Register host and guest users with the private channel.
			WebsocketServerInstance.PrivateChannels[newChannelId] = PrivateChannelInfo{client.user.Id, otherClient.user.Id}

			// Subscribe both host and guest users to the new channel.
			client.addChannelSubscription(newChannelId)
			otherClient.addChannelSubscription(newChannelId)

			result := WebsocketPayload{
				Status:    handlerStatusOk,
				ChannelId: newChannelId.String(),
				Category:  CategoryPrivate,
			}

			err = client.sendResponseMessage(websocketMessage, result)

			// Notify that user joined channel.
			notifyUserJoinedChannel(newChannelId, client.user)
			notifyUserJoinedChannel(newChannelId, otherClient.user)

			return err
		}
	}
	//endregion subscribe to a private channel

	//region subscribe to a non-cached space channel
	space, err := models.GetSpaceById(WebsocketServerInstance.Db, *channelId)
	if space != nil {

		//region server channel cache
		var exists = false
		for _, v := range WebsocketServerInstance.ServerChannels {
			if v == space.Id {
				exists = true
				break
			}
		}

		if !exists {
			WebsocketServerInstance.SpaceChannels = append(WebsocketServerInstance.SpaceChannels, space.Id)
		}
		//endregion server channel cache

		//region subscription
		client.addChannelSubscription(*channelId)
		//endregion subscription

		//region response
		result := WebsocketPayload{
			Status:    handlerStatusOk,
			ChannelId: channelId.String(),
			Category:  CategorySpace,
		}
		err = client.sendResponseMessage(websocketMessage, result)
		//endregion response

		//region notify
		notifyUserJoinedChannel(*channelId, client.user)
		//endregion notify

		//region presence
		err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, PresenceStatusAvailable, *channelId, client.user.Presence.ServerId)
		if err != nil {
			log.Errorf(err.Error())
			result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
			return client.sendResponseMessage(websocketMessage, result)
		}

		err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)
		//endregion presence

		return err
	}
	//endregion subscribe to a non-cached space channel

	//region subscribe to a non-cached server channel
	server, err := models.GetServerById(WebsocketServerInstance.Db, *channelId)
	if server != nil {
		//region server channel cache
		var exists = false
		for _, v := range WebsocketServerInstance.ServerChannels {
			if v == server.Id {
				exists = true
				break
			}
		}

		if !exists {
			WebsocketServerInstance.ServerChannels = append(WebsocketServerInstance.ServerChannels, server.Id)
		}
		//endregion server channel cache

		//region subscribe
		client.addChannelSubscription(*channelId)
		//endregion subscribe

		//region response
		result := WebsocketPayload{
			Status:    handlerStatusOk,
			ChannelId: channelId.String(),
			Category:  CategoryServer,
		}

		err = client.sendResponseMessage(websocketMessage, result)
		//endregion response

		//region notify
		notifyUserJoinedChannel(*channelId, client.user)
		//endregion notify

		//region presence
		err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, PresenceStatusAvailable, client.user.Presence.SpaceId, *channelId)
		if err != nil {
			log.Errorf(err.Error())
			result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
			return client.sendResponseMessage(websocketMessage, result)
		}

		err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)
		//endregion presence

		return err
	}
	//endregion subscribe to a non-cached server channel

	// Disallow to join arbitrary channel.
	err = fmt.Errorf("channel {%s} does not exist", channelId)
	log.Errorf(err.Error())
	result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
	return client.sendResponseMessage(websocketMessage, result)
}

func channelUnsubscribeHandler(client *WebsocketClient, websocketMessage *WebsocketMessage, _ WebsocketTopic, _ string, args interface{}) (err error) {

	// Parse request args as the string key map.
	m, ok := args.(map[string]interface{})
	if !ok {
		err := fmt.Sprintf("unable to parse args: %+v", args)
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}

	userId, err := getUserId(m)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	if client.user == nil || client.user.Id != userId {
		err := fmt.Sprintf("client is not authorized")
		log.Errorf(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}

	// Get the channel the message is sent to.
	channelId, err := getChannelId(m)
	if err != nil {
		log.Errorf(err.Error())
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	notifyUserLeftChannel(*channelId, client.user)
	client.removeChannelSubscription(*channelId)

	//region update user presence
	if len(client.channels) == 0 || WebsocketServerInstance.SystemChannel == *channelId {
		err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, PresenceStatusOffline, uuid.Nil, uuid.Nil)
		if err != nil {
			log.Errorf(err.Error())
			result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
			return client.sendResponseMessage(websocketMessage, result)
		}

		err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)
		if err != nil {
			log.Errorf(err.Error())
			result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
			return client.sendResponseMessage(websocketMessage, result)
		}
	} else {
		var found = false
		for _, c := range WebsocketServerInstance.SpaceChannels {
			if c == *channelId {
				err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, PresenceStatusAvailable, uuid.Nil, client.user.Presence.ServerId)
				if err != nil {
					log.Errorf(err.Error())
					result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
					return client.sendResponseMessage(websocketMessage, result)
				}

				err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)
				if err != nil {
					log.Errorf(err.Error())
					result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
					return client.sendResponseMessage(websocketMessage, result)
				}

				found = true
				break
			}
		}

		if !found {
			for _, c := range WebsocketServerInstance.ServerChannels {
				if c == *channelId {
					err = client.user.UpdateUserPresence(WebsocketServerInstance.Db, PresenceStatusAvailable, client.user.Presence.SpaceId, uuid.Nil)
					if err != nil {
						log.Errorf(err.Error())
						result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
						return client.sendResponseMessage(websocketMessage, result)
					}

					err = notifyUserPresenceChanged(WebsocketServerInstance.SystemChannel, client.user)
					if err != nil {
						log.Errorf(err.Error())
						result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
						return client.sendResponseMessage(websocketMessage, result)
					}

					found = true
					break
				}
			}
		}
	}
	//endregion

	result := WebsocketPayload{
		Status:    handlerStatusOk,
		ChannelId: channelId.String(),
		Category:  getCategoryByChannelId(channelId),
	}
	return client.sendResponseMessage(websocketMessage, result)
}

func vivoxHandler(client *WebsocketClient, websocketMessage *WebsocketMessage, topic WebsocketTopic, method string, args interface{}) (err error) {
	if topic != VivoxTopic {
		err := fmt.Errorf("wrong topic %d for the method %s", topic, method)
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	m, ok := args.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("unable to parse args %s", args)
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	userId, err := getUserId(m)
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	if client.user == nil || client.user.Id != userId {
		err := fmt.Sprintf("client is not authenticated")
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}

	jsonBody, err := json.Marshal(m[handlerArgsVivoxPayload])
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	var tokenPayload models.VivoxTokenPayload

	err = json.Unmarshal(jsonBody, &tokenPayload)
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	var jsonPayload string

	if method == VivoxGetLoginTokenMethod {
		jsonPayload, err = models.GetLoginToken(tokenPayload)
	} else if method == VivoxGetJoinTokenMethod {
		jsonPayload, err = models.GetJoinToken(tokenPayload)
	} else if method == VivoxMuteMethod {
		jsonPayload, err = models.RequestMute(tokenPayload)
	} else if method == VivoxUnmuteMethod {
		jsonPayload, err = models.RequestUnmute(tokenPayload)
	} else if method == VivoxKickMethod {
		jsonPayload, err = models.RequestKick(tokenPayload)
	}

	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	result := WebsocketPayload{Status: handlerStatusOk, Message: jsonPayload}
	return client.sendResponseMessage(websocketMessage, result)
}

func userActionHandler(client *WebsocketClient, websocketMessage *WebsocketMessage, _ WebsocketTopic, _ string, args interface{}) (err error) {
	m, ok := args.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("unable to parse args %s", args)
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	userId, err := getUserId(m)
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	if client.user == nil || client.user.Id != userId {
		err := fmt.Sprintf("client is not authenticated")
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err}
		return client.sendResponseMessage(websocketMessage, result)
	}

	jsonBody, err := json.Marshal(m[handlerArgsMessage])
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	var action models.Action

	err = json.Unmarshal(jsonBody, &action)
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	err = models.AddAction(WebsocketServerInstance.Db, action)
	if err != nil {
		log.Error(err)
		result := WebsocketPayload{Status: handlerStatusError, Message: err.Error()}
		return client.sendResponseMessage(websocketMessage, result)
	}

	result := WebsocketPayload{Status: handlerStatusOk}
	return client.sendResponseMessage(websocketMessage, result)
}

//region broadcast and notify helpers

func broadcastMessageToChannel(channelId uuid.UUID, payload WebsocketPayload) {
	// Broadcast the message to channels.
	for _, client := range WebsocketServerInstance.Clients {
		// Check if client is subscribed for the channel.
		if bSubscribed := containsUUID(client.channels, channelId); bSubscribed {
			// Broadcast the message to subscribed clients.
			if err := client.SendPushMessage(ChatTopic, payload); err != nil {
				log.Errorf("got an error sending push message to websocket client {%s}", client.Id)
			}
		}
	}
}

func multicastMessageToChannel(userIds []uuid.UUID, channelId uuid.UUID, payload WebsocketPayload) {
	// Broadcast the message to channels.
	for _, client := range WebsocketServerInstance.Clients {
		// Check if client is subscribed for the channel.
		if bSubscribed := containsUUID(client.channels, channelId); bSubscribed {
			if bShouldSend := containsUUID(userIds, client.user.Id); bShouldSend {
				// Multicast the message to subscribed clients.
				if err := client.SendPushMessage(ChatTopic, payload); err != nil {
					log.Errorf("got an error sending push message to websocket client {%s}", client.Id)
				}
			}
		}
	}
}

func notifyUserPresenceChanged(channelId uuid.UUID, user *models.User) (err error) {

	jsonPayload, err := json.Marshal(user.Presence)
	if err != nil {
		log.Error(err)
		return err
	}

	payload := WebsocketPayload{
		Status:    handlerStatusOk,
		Message:   string(jsonPayload),
		Sender:    user,
		ChannelId: channelId.String(),
		Category:  getCategoryByChannelId(&channelId),
	}

	leaders, err := models.GetUserLeadersById(WebsocketServerInstance.Db, user.Id)
	if err != nil {
		log.Error(err)
		return err
	}

	var userIds []uuid.UUID
	for _, v := range leaders {
		userIds = append(userIds, v.Id)
	}

	multicastMessageToChannel(userIds, channelId, payload)

	return nil
}

func notifyUserJoinedChannel(channelId uuid.UUID, user *models.User) {
	payload := WebsocketPayload{
		Status:    handlerStatusOk,
		Message:   MessageNotifyUserJoinedChannel,
		Sender:    user,
		ChannelId: channelId.String(),
		Category:  getCategoryByChannelId(&channelId),
	}

	broadcastMessageToChannel(channelId, payload)
}

func notifyUserLeftChannel(channelId uuid.UUID, user *models.User) {
	payload := WebsocketPayload{
		Status:    handlerStatusOk,
		Message:   MessageNotifyUserLeftChannel,
		Sender:    user,
		ChannelId: channelId.String(),
		Category:  getCategoryByChannelId(&channelId),
	}

	broadcastMessageToChannel(channelId, payload)
}

//endregion notify helpers

//region get and find helpers

func findExistingPrivateChannelForUsers(userId uuid.UUID, otherUserId uuid.UUID) *uuid.UUID {
	for privateChannelId, privateChannel := range WebsocketServerInstance.PrivateChannels {
		if privateChannel.Host == userId && privateChannel.Guest == otherUserId {
			return &privateChannelId
		}
		if privateChannel.Host == otherUserId && privateChannel.Guest == userId {
			return &privateChannelId
		}
	}
	return nil
}

func getChannelId(args map[string]interface{}) (*uuid.UUID, error) {

	// Get the channel the message is sent to.
	channelIdStr, ok := args[handlerArgsChannel].(string)
	if !ok {
		return nil, fmt.Errorf("can not parse channel id as string: %s", args[handlerArgsChannel])
	}

	// Check that the channel ID is correct.
	channelId, err := uuid.Parse(channelIdStr)
	if err != nil {
		return nil, fmt.Errorf("can not parse channel UUID: %s", channelIdStr)
	}

	return &channelId, nil
}

func getUserId(args map[string]interface{}) (uuid.UUID, error) {

	userIdStr, ok := args[handlerArgsUserId].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("can not parse user api key as string: %s", args[handlerArgsUserId])
	}

	return uuid.Parse(userIdStr)
}

func getCategoryByChannelId(channelId *uuid.UUID) string {
	if WebsocketServerInstance.SystemChannel == *channelId {
		return CategorySystem
	}

	if WebsocketServerInstance.GeneralChannel == *channelId {
		return CategoryGeneral
	}

	for _, c := range WebsocketServerInstance.SpaceChannels {
		if c == *channelId {
			return CategorySpace
		}
	}

	for _, c := range WebsocketServerInstance.ServerChannels {
		if c == *channelId {
			return CategoryServer
		}
	}

	for c := range WebsocketServerInstance.PrivateChannels {
		if c == *channelId {
			return CategoryPrivate
		}
	}

	return CategoryUnknown
}

//endregion getter helpers

//region authentication helpers

func registerSender(client *WebsocketClient, id uuid.UUID) error {
	user, err := models.GetUserById(WebsocketServerInstance.Db, id)

	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	client.user = user

	return nil
}

//endregion authentication helpers
