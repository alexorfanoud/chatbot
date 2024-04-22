package handlers

import "chat/internal/conversation"

var conversationManager = conversation.ConversationManagerImpl{}
var notifier = conversation.WebsocketConnectionNotifier{}
