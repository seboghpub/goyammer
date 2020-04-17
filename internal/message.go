package internal

import (
	"fmt"
	"strconv"
)

// Message is the data structure to represent a set of messages.
type Message struct {
	YammerMessage
}

// Messages is the data structure to represent all messages.
type Messages struct {
	client *Client

	// all messages by message id
	cache map[int64]*Message

	// message list by group id (-1 indicates private messages)
	messageLists map[int64][]*Message

	// id of the latest message by group id (-1 indicates private messages)
	latest map[int64]int64
}

// NewMessages returns a new Messages object.
func NewMessages(client *Client) *Messages {
	return &Messages{
		client:       client,
		cache:        make(map[int64]*Message),
		messageLists: make(map[int64][]*Message),
		latest:       make(map[int64]int64),
	}
}

// GetNewMessages returns new messages for the given group (in chronological order).
func (messages *Messages) GetNewMessages(groupId int64) ([]*Message, error) {

	// construct path (private by default, for a particular group if groupId is !-1)
	path := "messages/private.json"
	if groupId != -1 {
		path = fmt.Sprintf("messages/in_group/%d.json", groupId)
	}

	// if we don't have a latest id, get one and return
	if _, ok := messages.latest[groupId]; !ok {

		// construct parameters
		params := map[string]string{"limit": "1"}

		// construct request
		req, errReq := messages.client.newRequest("GET", path, params, nil)
		if errReq != nil {
			return nil, fmt.Errorf("failed to construct latest request for group %d: %v", groupId, errReq)
		}

		// do request and parse response
		var ymr YammerMessageResponse
		_, errDo := messages.client.do(req, &ymr)
		if errDo != nil {
			return nil, fmt.Errorf("failed to do latest request for group %d: %v", groupId, errDo)
		}

		// set the latest
		messages.latest[groupId] = ymr.Messages[0].ID
		return []*Message{}, nil
	}

	// construct parameters
	params := map[string]string{"newer_than": strconv.FormatInt(messages.latest[groupId], 10)}

	// construct request
	req, errReq := messages.client.newRequest("GET", path, params, nil)
	if errReq != nil {
		return nil, fmt.Errorf("failed to construct messages request for group %d: %v", groupId, errReq)
	}

	// do request and parse response
	var ymr YammerMessageResponse
	_, errDo := messages.client.do(req, &ymr)
	if errDo != nil {
		return nil, fmt.Errorf("failed to do latest request for group %d: %v", groupId, errDo)
	}

	// count messages
	messageCount := len(ymr.Messages)

	// return if no new messages
	if len(ymr.Messages) < 1 {
		return []*Message{}, nil
	}

	// extract messages ids and cache messages
	var newMessages []*Message = make([]*Message, int(messageCount))
	for i := 0; i < messageCount; i++ {
		yammerMessage := ymr.Messages[i]

		message := &Message{yammerMessage}

		// extract yammerMessage id
		newMessages[messageCount-i-1] = message

		// store yammerMessage in the cache
		messages.cache[yammerMessage.ID] = message
	}

	// update latest id
	messages.latest[groupId] = ymr.Messages[0].ID

	// update message lists
	if messageList, ok := messages.messageLists[groupId]; ok {
		messages.messageLists[groupId] = append(messageList, newMessages...)
	} else {
		messages.messageLists[groupId] = newMessages
	}

	return newMessages, nil
}
