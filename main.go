package main

import (
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	token := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	if token == "" {
		log.Fatal("SLACK_BOT_TOKEN environment variable not set")
	}
	if appToken == "" {
		log.Fatal("SLACK_APP_TOKEN environment variable not set (should start with xapp-)")
	}

	api := slack.New(token, slack.OptionAppLevelToken(appToken))
	socketClient := socketmode.New(api)

	// Precheck: verify bot authentication
	authTest, err := api.AuthTest()
	if err != nil {
		log.Fatalf("Slack authentication failed: %v", err)
	}
	log.Printf("Logged in as bot user: %s (ID: %s)", authTest.User, authTest.UserID)

	go func() {
		for evt := range socketClient.Events {
			switch evt.Type {
			case socketmode.EventTypeEventsAPI:
				event, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Ignored %+v\n", evt)
					continue
				}
				socketClient.Ack(*evt.Request)
				if event.Type == slackevents.CallbackEvent {
					innerEvent := event.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						handleMessageEvent(ev, api, authTest.UserID)
					}
				}
			}
		}
	}()

	log.Println("Mirror Bot is running in Socket Mode...")
	socketClient.Run()
}

func handleMessageEvent(ev *slackevents.MessageEvent, api *slack.Client, botUserID string) {
	// Print event info to the console for debugging
	if containsMention(ev.Text, botUserID) {
		fmt.Printf("Received MessageEvent: %+v\n", *ev)
	}
}

func containsMention(text, botUserID string) bool {
	return (botUserID != "" && (containsIgnoreCase(text, "@mirrorbot") || containsIgnoreCase(text, "<@"+botUserID+">")))
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && (containsIgnoreCase(s[1:], substr) || containsIgnoreCase(s, substr[1:])))) || (len(s) > 0 && len(substr) > 0 && (s[0]|32) == (substr[0]|32) && containsIgnoreCase(s[1:], substr[1:]))
}
