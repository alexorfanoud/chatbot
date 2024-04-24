package notification

import (
	"chat/internal/data/store/memory"
	"chat/internal/utils"
	"context"
	"fmt"
	"strconv"
)

var sessionStore = memory.InMemorySessionStore{}

type WebChannelNotifier struct{}

func (*WebChannelNotifier) Notify(ctx context.Context, userId int64, msg string) error {
	session := sessionStore.Get(strconv.FormatInt(userId, 10))
	if session == nil {
		utils.Log(ctx, fmt.Sprintf("Unable to find session for user id: %d, ignoring...", userId))
		return nil
	}

	err := session.WriteMessage(1, []byte(msg))
	return err
}

func (*WebChannelNotifier) NotifyStream(ctx context.Context, userId int64, stream <-chan string) error {
	session := sessionStore.Get(strconv.FormatInt(userId, 10))
	if session == nil {
		utils.Log(ctx, fmt.Sprintf("Unable to find session for user id: %d, ignoring...", userId))
		return nil
	}

	for {
		data, ok := <-stream
		if !ok {
			break
		}

		if err := session.WriteMessage(1, []byte(data)); err != nil {
			utils.Log(ctx, fmt.Sprintf("Unable to send message over websocket connection: %s", err.Error()))
			return err
		}
	}
	return nil
}
