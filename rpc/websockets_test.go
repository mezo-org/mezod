package rpc

import (
	"net/http/httptest"
	"strings"
	"testing"

	"cosmossdk.io/log"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestWebsocketsServerReadLimit(t *testing.T) {
	server := httptest.NewServer(&websocketsServer{logger: log.NewNopLogger()})
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	t.Run("message within the limit is read", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Not valid JSON, so the server reads it and responds with an error.
		message := make([]byte, wsReadLimit)
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, message))

		_, response, err := conn.ReadMessage()
		require.NoError(t, err)
		require.Contains(t, string(response), "error")
	})

	t.Run("message over the limit closes the connection", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		require.NoError(t, err)
		defer conn.Close()

		message := make([]byte, wsReadLimit+1)
		// The server may close the connection before the whole message is
		// written, so the write error is not checked.
		_ = conn.WriteMessage(websocket.TextMessage, message)

		_, _, err = conn.ReadMessage()
		require.Error(t, err)
	})
}
