package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sebrandon1/mirrorbot/pkg/ocpmirror"
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
	// Ignore messages sent by the bot itself
	if ev.User == botUserID {
		return
	}
	// Print event info to the console for debugging
	if containsMention(ev.Text, botUserID) {
		fmt.Printf("Received MessageEvent: %+v\n", *ev)
		// Parse for OCP version (e.g., "4.20")
		fields := strings.Fields(ev.Text)
		var version string
		for _, f := range fields {
			if strings.Count(f, ".") == 1 && strings.HasPrefix(f, "4.") {
				version = f
				break
			}
		}
		if version != "" {
			releases, err := ocpmirror.ListReleases(version)
			if err != nil {
				msg := fmt.Sprintf("Error fetching releases for %s: %v", version, err)
				fmt.Println(msg)
				_, _, _ = api.PostMessage(ev.Channel, slack.MsgOptionText(msg, false))
				return
			}
			if len(releases) == 0 {
				msg := fmt.Sprintf("No releases found for %s", version)
				fmt.Println(msg)
				_, _, _ = api.PostMessage(ev.Channel, slack.MsgOptionText(msg, false))
				return
			}
			latest := releases[0]

			// Fetch release status and pullSpec from status API
			status, err := ocpmirror.FetchReleaseStatus(latest.Version)
			if err != nil {
				fmt.Printf("Warning: could not fetch status for %s: %v\n", latest.Version, err)
			}
			detail, err := ocpmirror.FetchReleaseDetail(latest.Version)
			if err != nil {
				fmt.Printf("Warning: could not fetch pullSpec for %s: %v\n", latest.Version, err)
			}

			msg := fmt.Sprintf(
				"Latest %s release in %s: %s\nURL: %s",
				version, latest.Folder, latest.Version, latest.URL,
			)
			// Determine which stream was used for the release status
			// Try to infer the correct stream for the detail page link
			stream := latest.Folder
			if stream == "ocp-dev-preview" {
				stream = "4-dev-preview"
			} else if stream == "ocp" {
				stream = "4-stable"
			}
			if status != nil {
				createdTime, err := time.Parse(time.RFC3339, status.Created)
				if err == nil {
					daysAgo := int(time.Since(createdTime).Hours() / 24)
					msg += fmt.Sprintf("\nCreated: %d days ago (%s)", daysAgo, status.Created)
				} else {
					msg += fmt.Sprintf("\nCreated: %s", status.Created)
				}
				msg += fmt.Sprintf("\nPhase: %s", status.Phase)
				if status.KubernetesVersion != "" {
					kubeVer := status.KubernetesVersion
					kubeVerParts := strings.SplitN(kubeVer, ".", 3)
					if len(kubeVerParts) >= 2 {
						major := kubeVerParts[0]
						minor := kubeVerParts[1]
						kubeReleaseURL := fmt.Sprintf("https://kubernetes.io/releases/#release-v%s-%s", major, minor)
						msg += fmt.Sprintf("\nKubernetes Version: <%s|%s>", kubeReleaseURL, kubeVer)
					} else {
						msg += fmt.Sprintf("\nKubernetes Version: %s", kubeVer)
					}
				}
				if status.RHCOSVersion != "" {
					if status.RHCOSFrom != "" {
						msg += fmt.Sprintf("\nRHCOS Version: %s (%s)", status.RHCOSVersion, status.RHCOSFrom)
					} else {
						msg += fmt.Sprintf("\nRHCOS Version: %s", status.RHCOSVersion)
					}
				}
				// Add clickable link to release detail page
				msg += fmt.Sprintf("\nClick <https://openshift-release.apps.ci.l2s4.p1.openshiftapps.com/releasestream/%s/release/%s|here> for release info", stream, latest.Version)
			}
			if detail != nil && detail.PullSpec != "" {
				msg += fmt.Sprintf("\nInstall: oc adm release extract --command=oc --from=%s", detail.PullSpec)
			}
			msg = ">```" + msg + "```"
			fmt.Println(msg)
			api.PostMessage(
				ev.Channel,
				slack.MsgOptionText(msg, false),
				slack.MsgOptionDisableLinkUnfurl(),
				slack.MsgOptionDisableMediaUnfurl(),
			)
		}
	}
}

func containsMention(text, botUserID string) bool {
	return (botUserID != "" && (containsIgnoreCase(text, "@mirrorbot") || containsIgnoreCase(text, "<@"+botUserID+">")))
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && (containsIgnoreCase(s[1:], substr) || containsIgnoreCase(s, substr[1:])))) || (len(s) > 0 && len(substr) > 0 && (s[0]|32) == (substr[0]|32) && containsIgnoreCase(s[1:], substr[1:]))
}
