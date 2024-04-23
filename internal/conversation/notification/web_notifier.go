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
