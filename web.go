package slacker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// the base URL for the Slack web API
const BaseAPI = "https://slack.com/api/"

// the RTM start method
const RtmStart = "rtm.start"

// Slack can be used to send API calls to the Slack API, and also
// maintains an RTM (real-time messaging) connection.
type Slack struct {
	apiToken string
	rtm      *Rtm
}

// RtmStartResponse is the JSON response from the rtm.start method.
type RtmStartResponse struct {
	OK       bool        `json:"ok"`
	Err      interface{} `json:"error"`
	URL      string      `json:"url"`
	Self     *User       `json:"self"`
	Team     *Team       `json:"team"`
	Users    []User      `json:"users"`
	Channels []Channel   `json:"channels"`
	Groups   []Group     `json:"groups"`
	IMs      []IM        `json:"ims"`
	Bots     []BotInfo   `json:"bots"`
}

// User is the Slack API's user response object.
type User struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Prefs          map[string]interface{} `json:"prefs"`
	Created        int64                  `json:"created"`
	ManualPresence string                 `json:"manual_presence"`
}

// Team is the Slack API's team response object.
type Team struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	EmailDomain       string                 `json:"email_domain"`
	Domain            string                 `json:"domain"`
	MsgEditWindowMins int64                  `json:"msg_edit_window_mins"`
	OverStorageLimit  bool                   `json:"over_storage_limit"`
	Prefs             map[string]interface{} `json:"prefs"`
}

// Channel is the Slack API's channel response object.
type Channel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IsChannel   bool      `json:"is_channel"`
	Created     int64     `json:"created"`
	Creator     string    `json:"creator"`
	IsArchived  bool      `json:"is_archived"`
	IsGeneral   bool      `json:"is_general"`
	Members     []string  `json:"members"`
	Topic       *SetValue `json:"topic"`
	Purpose     *SetValue `json:"purpose"`
	IsMember    bool      `json:"is_member"`
	LastRead    string    `json:"last_read"`
	Latest      *Message  `json:"latest"`
	UnreadCount int64     `json:"unread_count"`
}

// SetValue is a type used by Slack to represent a string value that
// has a creater and a set date, e.g., the topic.
type SetValue struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

// Group is a placeholder for now.
type Group struct {
}

// IM is a placeholder for now.
type IM struct {
}

// BotInfo is a placeholder for now.
type BotInfo struct {
}

// Message is a basic map that is used for JSON messages with the Slack APIs.
type Message map[string]interface{}

// post send a POST to the Slack Web API
func (s *Slack) post(endpoint string, params map[string]string, value interface{}) error {
	values := map[string][]string{}
	for k, v := range params {
		values[k] = []string{v}
	}
	values["token"] = []string{s.apiToken}
	resp, err := http.PostForm(BaseAPI+endpoint, url.Values(values))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Slack returned non-200 status code of %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, value)
}

// connect starts up an RTM connection
func (s *Slack) connect() error {
	var rtmResp RtmStartResponse
	err := s.post(RtmStart, map[string]string{}, &rtmResp)
	if err != nil {
		return err
	}

	if !rtmResp.OK {
		return fmt.Errorf("Error connecting to Slack: %s", rtmResp.Err)
	}

	url, err := url.Parse(rtmResp.URL)
	if err != nil {
		return err
	}
	// we require a port
	if !strings.Contains(url.Host, ":") {
		if url.Scheme == "wss" {
			url.Host += ":443"
		} else {
			url.Host += ":80"
		}
	}

	s.rtm, err = RtmConnect(url.String())
	return err
}

// Connect uses the given token to create a Slack API object, and starts
// up a RTM connection.
func Connect(token string) (*Slack, error) {
	slack := &Slack{token, nil}
	err := slack.connect()
	if err != nil {
		return nil, err
	}
	return slack, nil
}
