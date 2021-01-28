package ws

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
)

type Client struct {
	conn     *websocket.Conn
	RemoteIp string
	Send     chan Message
	Recv     chan Message
	closed   bool
}

type Message struct {
	MessageType int
	Data        []byte
}

func (c *Client) Close() {
	if !c.closed {
		c.closed = true
		c.conn.Close()
		close(c.Recv)
	}
}

func (c *Client) readPump() {
	defer c.Close()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Read a message from the agent.
		t, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Unexpected error while reading agent message:", err)
			}
			break
		}

		if c.closed {
			break
		}

		c.Recv <- Message{
			MessageType: t,
			Data:        msg,
		}
	}
}

func (c *Client) writePump() {
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		c.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Println("Client send channel closed")
				return
			}

			err := c.conn.WriteMessage(msg.MessageType, msg.Data)
			if err != nil {
				return
			}
		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}
