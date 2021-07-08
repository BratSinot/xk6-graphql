package graphql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/js/modules/k6/ws"
)

func init() {
	modules.Register("k6/x/graphql", new(GraphQL))
}

type GraphQL struct{}

type Client struct {
	rt         *goja.Runtime
	socket     *ws.Socket
	id         string
	logEnabled bool
}

func (c *GraphQL) XStart(ctx *context.Context, url string, initPayload interface{}, query string, logEnabled bool) (*ws.WSHTTPResponse, error) {
	rt := common.GetRuntime(*ctx)

	return ws.
		New().
		Connect(
			*ctx,
			url,
			rt.ToValue(struct {
				Headers map[string]string `json:"headers"`
			}{
				Headers: map[string]string{
					"sec-websocket-protocol": "graphql-transport-ws",
				},
			}),
			rt.ToValue(
				func(socket *ws.Socket) {
					client := Client{
						rt,
						socket,
						uuid.NewString(),
						logEnabled,
					}
					client.start(initPayload, query)
				},
			))
}

func (c *Client) log(msg string) {
	if c.logEnabled {
		fmt.Println(msg)
	}
}

func (c *Client) send(data map[string]interface{}) {
	msg, _ := json.Marshal(data)

	c.log(fmt.Sprintf("Sending: %s", msg))
	c.socket.Send(string(msg))
}

func (c *Client) start(initPayload interface{}, query string) {
	c.socket.On("open", c.rt.ToValue(func() {
		c.log("Connected.")

		c.send(map[string]interface{}{
			"type":    "connection_init",
			"payload": initPayload,
		})
	}))

	c.socket.On("message", c.rt.ToValue(func(data string) {
		if len(data) == 0 {
			return
		}

		objs := map[string]interface{}{}
		if err := json.Unmarshal([]byte(data), &objs); err == nil {
			switch t := objs["type"]; t {
			case "connection_ack":
				c.subscribe(query)
			case "next":
				c.log(fmt.Sprintf("Got message: %v", objs["payload"]))
			default:
				c.log(fmt.Sprint("Empty or unknown type: ", t))
			}
		} else {
			c.log(fmt.Sprintf("Got error: %v", err))
		}
	}))

	c.socket.On("error", c.rt.ToValue(func(err error) {
		if err != nil && err.Error() != "websocket: close sent" {
			c.log(fmt.Sprintf("An unexpected error occurred: %s", err))
		}
	}))

	c.socket.On("close", c.rt.ToValue(func() {
		c.log("Disconnected.")
	}))
}

func (c *Client) subscribe(query string) {
	c.send(map[string]interface{}{
		"id":   c.id,
		"type": "start",
		"payload": map[string]string{
			"query": query,
		},
	})
}
