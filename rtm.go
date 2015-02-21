package slacker

import (
	"fmt"
	"sync/atomic"
	"time"

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
	error          bool
}

// RtmConnect connects to the given URL, which should be an authenticated URL
// from the rtm.start API call.
func RtmConnect(slack *Slack, url string) (*Rtm, error) {
	ws, err := websocket.Dial(url, "", "http://localhost")
	if err != nil {
		return nil, err
	}

	rtm := &Rtm{ws, 1, false}

	// start ping keep-aliver
	go func() {
		tick := time.Tick(30 * time.Second)
		for {
			<-tick
			slack.out <- map[string]interface{}{"type": "ping"}
		}
	}()

	go func() {
		for {
			message := Message{}
			err := websocket.JSON.Receive(rtm.ws, &message)
			if err != nil {
				fmt.Printf("Error in websocket receive: %s\n", err.Error())
				rtm.error = true
				return
			}
			slack.in <- message
		}
	}()

	go func() {
		for {
			message := <-slack.out
			counter := atomic.AddInt64(&rtm.messageCounter, 1) - 1
			message["id"] = counter
			fmt.Printf("Sending %v\n", message)
			err := websocket.JSON.Send(rtm.ws, message)
			if err != nil {
				fmt.Printf("Error in websocket receive: %s", err.Error())
				rtm.error = true
				slack.out <- message
				return
			}
		}
	}()

	return rtm, nil
}
