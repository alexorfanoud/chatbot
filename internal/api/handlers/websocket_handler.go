package handlers

import (
	"chat/internal/api/middleware"
	"chat/internal/conversation/notification"
	"chat/internal/data/store/distrib"
	"chat/internal/data/store/memory"
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var sessionStore = memory.InMemorySessionStore{}

func HandleWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}

	connCtx := context.Background()
	userId := r.URL.Query().Get("user")
	sessionStore.Save(userId, conn)
	uidi, err := strconv.ParseInt(userId, 10, 64)
	distrib.RemoveSession(connCtx, uidi, model.WEB)
	utils.Log(connCtx, fmt.Sprintf("Received new session for user id: %s", userId))

	defer func() {
		conn.Close()
		sessionStore.Clear(userId)
	}()

	for {
		// For each incoming message
		ctx := context.Background()
		ctx, span := middleware.Tracer.Start(ctx, "handleMessage")
		defer span.End()
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Send it to the LLM for conversation processing
		sessionIdi, _ := strconv.ParseInt(userId, 10, 64)
		req := model.ConversationRequest{
			Request:                  string(p),
			UserID:                   sessionIdi,
			ChannelType:              model.WEB,
			WorkflowExecutionContext: &model.WorkflowExecutionContext{Workflow: model.UNKNOWN, ContextWindow: 5},
			NewConversation:          false}
		resCh, err := conversationManager.HandleRequestStreaming(ctx, req)
		if err != nil {
			log.Println("Error handling request:", err)
			break
		}

		// Stream response to the web client
		err = notification.GetNotifier(model.WEB).NotifyStream(ctx, sessionIdi, resCh)
		if err != nil {
			utils.Log(ctx, fmt.Sprintf("Error during notification: %s", err.Error()))
			break
		}
	}
}
