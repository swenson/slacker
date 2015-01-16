package main

import (
	"flag"
	"fmt"

	"github.com/swenson/slacker"
)

var token = flag.String("token", "", "Slack bot token to run with")

func main() {
	flag.Parse()
	if *token == "" {
		fmt.Printf("Must specify bot token with --token=your-token\n")
		return
	}
	bot, err := slacker.NewBot(*token)
	if err != nil {
		fmt.Printf("Error connecting to Slack: %s\n", err.Error())
		return
	}

	bot.RespondWith("whats up dog", func(user string, matchParts []string) string {
		return "not much, you?"
	})

	select {} // sleep forever
}
