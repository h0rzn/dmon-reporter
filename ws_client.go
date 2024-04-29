package main

import (
	"net/url"

	"github.com/gorilla/websocket"
)

const (
	addr = "localhost:4000"
	path = "/report_in/websocket"
)

type WSClient struct {
	con *websocket.Conn
}

func (c *WSClient) Connect() error {
	u := url.URL{Scheme: "ws", Host: addr, Path: path}

	con, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.con = con
	// defer c.con.Close()
	return nil
}

func (c *WSClient) Disconnect() error {
	return c.con.Close()
}

func (c *WSClient) Publish(data Metrics) error {
	// packet := make([]byte, 0)
	// _, err := c.con.Write(packet)
	// c.con.WriteJSON()
	return c.con.WriteJSON(data)
}
