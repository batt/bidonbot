package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/batt/bidonbot/slackbot"
	"github.com/nlopes/slack"
)

func loadFile(file_url string) map[string]string {
	m := make(map[string]string)
	resp, err := http.Get(file_url)
	if err != nil {
		log.Error(err)
		return nil
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		a := strings.Split(scanner.Text(), ":")
		if len(a) == 2 {
			k := strings.TrimSpace(a[0])
			v := strings.TrimSpace(a[1])
			m[k] = v
		} else {
			log.Warnf("Unable to split \"'%s\"", scanner.Text())
		}
	}
	return m
}

func main() {
	token := os.Getenv("SLACK_BOT_TOKEN")
	file_url := os.Getenv("RECYCLE_LIST")

	fmt.Printf("Recycle list url: %s\n\r", file_url)
	if token == "" {
		log.Fatalln("No slack token found!")
	}

	if file_url == "" {
		log.Fatalln("No recycle list found!")
	}

	// Slack Bot filter
	var opts slackbot.Config
	bot := slackbot.New(token, opts)

	recycle := loadFile(file_url)

	bot.DefaultResponse(func(b *slackbot.Bot, msg *slack.Msg) {
		reply := ""
		found := 0
		fmt.Printf("Message from channel (%s): %s\n", msg.Channel, msg.Text)
		trash := msg.Text
		for k, v := range recycle {
			if strings.Contains(strings.ToLower(k), strings.ToLower(trash)) {
				found++
				reply = reply + fmt.Sprintf("- %s -> %s\n", k, v)
			}
		}
		if found > 0 {
			bot.Message(msg.Channel, "Ho trovato i seguenti rifiuti:\n"+reply)
		} else {
			bot.Message(msg.Channel, "Non sono in grado di trovare nulla, sorry :(")
		}
	})

	//bot.RespondTo("^(.*)$", func(b *slackbot.Bot, msg *slack.Msg, args ...string) {
	//    fmt.Printf("Message from channel (%s): %s", msg.Channel, msg.Text)
	//    bot.Message(msg.Channel, "Antani la supercazzola, con scappellamento a destra!")
	//})

	fmt.Printf("Run Bot server\n\r")
	go func(b *slackbot.Bot) {
		if err := b.Start(); err != nil {
			log.Fatalln(err)
		}
	}(bot)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	fmt.Printf("Run HTTP server on port:%v\n\r", httpPort)
	http.ListenAndServe(":"+httpPort, nil)
}
