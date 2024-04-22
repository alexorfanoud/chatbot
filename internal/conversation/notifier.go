package conversation

import (
	"chat/internal/data/store/memory"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"errors"
	"fmt"
	"strconv"
)

type Notifier interface {
	Notify(context.Context, *model.User, string) error
}

type WebsocketConnectionNotifier struct{}

var sessionStore = memory.InMemorySessionStore{}

func (*WebsocketConnectionNotifier) Notify(ctx context.Context, user *model.User, msg string) error {
	session := sessionStore.Get(strconv.FormatInt(user.Id, 10))
	if session == nil {
		utils.Log(ctx, fmt.Sprintf("Unable to find session for user id: %d", user.Id))
		return errors.New(fmt.Sprintf("Unable to find session for userId: %d", user.Id))
	}

	err := session.WriteMessage(1, []byte(msg))
	return err
}
