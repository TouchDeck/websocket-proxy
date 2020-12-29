package main

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

type websocketClient struct {
	conn     *websocket.Conn
	remoteIp string
	send     chan message
	recv     chan message
	closed   bool
}

type message struct {
	messageType int
	data        []byte
}

func (c *websocketClient) close() {
	if !c.closed {
		c.closed = true
		c.conn.Close()
		close(c.recv)
	}
}

func (c *websocketClient) readPump() {
	defer c.close()

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

		c.recv <- message{
			messageType: t,
			data:        msg,
		}
	}
}

func (c *websocketClient) writePump() {
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		c.close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Println("Client send channel closed")
				return
			}

			err := c.conn.WriteMessage(msg.messageType, msg.data)
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
