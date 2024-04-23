package model

type ChannelType int

const (
	WEB ChannelType = iota
	SMS
	EMAIL
	VOICE
)

func (ct *ChannelType) SupportsStreaming() bool {
	return *ct == WEB
}
