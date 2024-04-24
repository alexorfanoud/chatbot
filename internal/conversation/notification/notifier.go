package notification

import (
	"chat/internal/model"
	"context"
)

var (
	webchannelNotifier = WebChannelNotifier{}
	notifiers          = map[model.ChannelType]Notifier{
		model.WEB: &webchannelNotifier,
	}
)

type Notifier interface {
	Notify(context.Context, int64, string) error
	NotifyStream(context.Context, int64, <-chan string) error
}

func GetNotifier(ct model.ChannelType) Notifier {
	return notifiers[ct]
}
