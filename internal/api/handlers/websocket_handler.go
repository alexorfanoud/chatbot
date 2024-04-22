package handlers

import (
	"chat/internal/api/middleware"
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
	uidi, err := strconv.Atoi(userId)
	distrib.RemoveSession(connCtx, uidi)
	utils.Log(connCtx, fmt.Sprintf("Received new session for user id: %s", userId))

	defer func() {
		conn.Close()
		sessionStore.Clear(userId)
	}()

	for {
		ctx := context.Background()
		ctx, span := middleware.Tracer.Start(ctx, "handleMessage")
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		sessionIdi, _ := strconv.ParseInt(userId, 10, 64)
		resp, err := conversationManager.HandleRequest(ctx, string(p), sessionIdi, false, &model.WorkflowExecutionContext{Workflow: model.UNKNOWN, ContextWindow: 5})
		if err != nil {
			log.Println("Error handling request:", err)
			break
		}

		// Echo message back to client
		err = conn.WriteMessage(messageType, []byte(resp))
		if err != nil {
			log.Println("Error sending message:", err)
			break
		}

		span.End()
	}
}
