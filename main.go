package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/resend/resend-go/v2"
)

const CONFIG_FILEPATH = "/etc/uptimego/apps-available"

func sendEmail(appName, appURL string) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Error("RESEND_API_KEY not set.")
		return
	}

	receiver := os.Getenv("RECEIVER_EMAIL")
	if receiver == "" {
		log.Error("RECEIVER_EMAIL not set.")
		return
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "UptimeGo <uptimego@shoto.at>",
		To:      []string{receiver},
		Html:    fmt.Sprintf("<strong>%s</strong> is down!<br><a href='%s'>Open app</a>", appName, appURL),
		Subject: "UptimeGo Alert!",
	}

	log.Infof("Sending email to: '%s'\n", receiver)
	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Error(err)
	} else {
		log.Infof("Email sent: %s\n", sent.Id)
	}
}

func checkUptime(appName, appURL string) bool {
	resp, err := http.Get(appURL)

	if resp != nil {
		resp.Body.Close()
	}

	return err == nil
}

func main() {
	file, err := os.Open(CONFIG_FILEPATH)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	stats, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, stats.Size())

	_, err = file.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(string(buf), "\n")
	for _, appEntry := range parts {
		if appEntry == "" {
			continue
		}
		appParts := strings.Split(appEntry, "@")
		if len(appParts) != 2 {
			log.Error("App entry must be of the format: 'app_name@url'")
		} else {
			running := checkUptime(appParts[0], appParts[1])
			if !running {
				log.Errorf("APP: '%s' is not running!\n", appParts[0])
				sendEmail(appParts[0], appParts[1])
			} else {
				log.Infof("APP: '%s' is up!\n", appParts[0])
			}
		}
	}
}
