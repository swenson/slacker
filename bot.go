package slacker

import (
	"fmt"
	"regexp"
)

// Bot is the basic type for defining a simple chat bot.
type Bot struct {
	slack *Slack
	rules []Rule
}

// Rule is a basic type representing bot a bot response rule.
type Rule struct {
	matchRe *regexp.Regexp
	exec    func(string, []string) string
}

// NewBot creates a new bot that connects to Slack with the given bot token.
// It will connect automatically and start processing responses in a goroutine.
func NewBot(token string) (*Bot, error) {
	slack, err := Connect(token)
	if err != nil {
		return nil, err
	}
	bot := &Bot{slack, []Rule{}}
	go bot.run()
	return bot, nil
}

// ClearResponses can be used to delete all of the rules that a bot knows about.
func (b *Bot) ClearResponses() {
	b.rules = b.rules[:0]
}

func (b *Bot) run() {
	for {
		message := <-b.slack.in
		fmt.Printf("Incoming %v\n", message)
		kind, ok := message["type"]
		if !ok {
			continue
		}
		if kind != "message" {
			continue
		}
		if message["text"] == nil || message["channel"] == nil || message["user"] == nil {
			continue
		}
		text := message["text"].(string)
		channel := message["channel"].(string)
		user := message["user"].(string)

		for _, rule := range b.rules {
			parts := rule.matchRe.FindStringSubmatch(text)
			if parts != nil {
				text = rule.exec(user, parts)
				if text != "" {
					b.slack.SayID(channel, text)
				}
			}
		}
	}
}

// GetUser gets information about a user (by ID).
func (b *Bot) GetUser(id string) (*User, error) {
	return b.slack.GetUser(id)
}

// RespondWith responds to the given regular expression, if it matches
// anywhere in the string, by executing the given function to construct
// a response. The exec arguments are the user id and the array of
// the regular expression groups.
func (b *Bot) RespondWith(re string, exec func(string, []string) string) {
	b.rules = append(b.rules, Rule{regexp.MustCompile(re), exec})
}

// Say posts a message to a given channel.  The channel can be a name or an ID.
func (b *Bot) Say(channel, message string) error {
	return b.slack.Say(channel, message)
}
