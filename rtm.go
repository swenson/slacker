package slacker

import (
	"fmt"
	"sync/atomic"

	"golang.org/x/net/websocket"
)

// The maximum RTM message size we are willing to consider.
// Although with general WebSockets this is effectively unlimited,
// Slack seems to cap messages around 16 MB.
const MaxMessageSize = 1 << 24

// Rtm is a structure that holds state for a real-time messaging
// WebSocket connection
type Rtm struct {
	ws             *websocket.Conn
	messageCounter int64
	in             chan Message
	out            chan Message
}

// Say sends the given message on the given channel.
// Note that the channel needs to be the channel id, and not the channel name.
func (rtm *Rtm) Say(channel, text string) {
	rtm.out <- Message(map[string]interface{}{"type": "message", "channel": channel, "text": text})
}

// RtmConnect connects to the given URL, which should be an authenticated URL
// from the rtm.start API call.
func RtmConnect(url string) (*Rtm, error) {
	ws, err := websocket.Dial(url, "", "http://localhost")
	if err != nil {
		return nil, err
	}

	rtm := &Rtm{ws, 1, make(chan Message), make(chan Message)}
	go func() {
		for {
			message := Message{}
			err := websocket.JSON.Receive(rtm.ws, &message)
			if err != nil {
				panic(err)
			}
			rtm.in <- message
		}
	}()

	go func() {
		for {
			message := <-rtm.out
			counter := atomic.AddInt64(&rtm.messageCounter, 1) - 1
			message["id"] = counter
			fmt.Printf("Sending %v\n", message)
			err := websocket.JSON.Send(rtm.ws, message)
			if err != nil {
				panic(err)
			}
		}
	}()

	return rtm, nil
}
