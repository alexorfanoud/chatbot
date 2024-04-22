package memory

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/patrickmn/go-cache"
)

type WebsocketConnectionStore interface {
	Save(sid string, conn *websocket.Conn)
	Get(sid string) *websocket.Conn
	Clear(sid string)
}

var store = cache.New(30*time.Minute, 10*time.Minute)

type InMemorySessionStore struct {
}

func (*InMemorySessionStore) Save(sid string, conn *websocket.Conn) {
	store.Set(sid, conn, -1)
}

func (*InMemorySessionStore) Get(sid string) *websocket.Conn {
	if val, found := store.Get(sid); found {
		return val.(*websocket.Conn)
	}

	return nil
}

func (*InMemorySessionStore) Clear(sid string) {
	store.Delete(sid)
}
