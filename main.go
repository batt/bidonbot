package main

import (
	"bufio"
	"fmt"
	"html"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/batt/bidonbot/slackbot"
	"github.com/nlopes/slack"
)

type Trash struct {
	Type   string
	Method string
	Note   string
}

func loadFile(file_url string) map[string]Trash {
	m := make(map[string]Trash)
	resp, err := http.Get(file_url)
	if err != nil {
		log.Error(err)
		return nil
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		a := strings.Split(scanner.Text(), ";")
		if len(a) == 4 {
			k := strings.TrimSpace(a[0])
			v := Trash{strings.TrimSpace(a[1]), strings.TrimSpace(a[2]), strings.TrimSpace(a[3])}
			m[k] = v
		} else {
			log.Warnf("Unable to split \"%s\"", scanner.Text())
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
		re := regexp.MustCompile("<@.*>")
		text := re.ReplaceAllString(msg.Text, "")
		fmt.Println(text)
		trash := strings.Split(strings.ToLower(text), " ")
		for k, v := range recycle {
			allFound := true
			for _, t := range trash {
				if !strings.Contains(strings.ToLower(k), t) {
					allFound = false
				}
			}
			if allFound {
				reply = reply + fmt.Sprintf("- %s -> %s", k, v.Type)
				if v.Method != "" {
					reply = reply + fmt.Sprintf(" conferire in: %s", v.Method)
				}
				if v.Note != "" {
					reply = reply + fmt.Sprintf(" - NOTE: %s", v.Note)
				}
				reply = reply + "\n"
				found++
			}
		}
		if found > 16 {
			bot.Message(msg.Channel, fmt.Sprintf("Ho trovato %d corrispondenze! Sii più preciso cribbio\n", found))
		} else if found > 0 {
			bot.Message(msg.Channel, "Ho trovato i seguenti rifiuti:\n"+reply)
		} else {
			bot.Message(msg.Channel, fmt.Sprintf("Non ho trovato niente che corrisponda a '%s', mi spiace :(", text))
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	httpURL := os.Getenv("HTTPURL")
	if httpURL == "" {
		httpURL = "https://bidonbot.herokuapp.com"
	}

	http.HandleFunc("/keepalive", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
		fmt.Println("keepalice pong")
	})

	wakeUpTime := os.Getenv("WAKEUP_TIME")
	if wakeUpTime == "" {
		wakeUpTime = "6:00"
	}

	sleepTime := os.Getenv("SLEEP_TIME")
	if sleepTime == "" {
		sleepTime = "21:00"
	}

	w := strings.Split(wakeUpTime, ":")
	s := strings.Split(sleepTime, ":")

	wh, _ := strconv.Atoi(w[0])
	wm, _ := strconv.Atoi(w[1])
	wakeUpOffset := (60*wh + wm) % (60 * 24)

	sh, _ := strconv.Atoi(s[0])
	sm, _ := strconv.Atoi(s[1])
	awakeMinutes := (60*(sh+24) + sm - wakeUpOffset) % (60 * 24)

	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			now := time.Now()

			elapsedMinutes := (60*(now.Hour()+24) + now.Minute() - wakeUpOffset) % (60 * 24)
			fmt.Printf("Awake for %d minutes\n", elapsedMinutes)
			if elapsedMinutes < awakeMinutes {
				fmt.Println("keepalive ping!")
				http.Get(httpURL + "/keepalive")
			} else {
				fmt.Println("skipping keepalive, going to sleep...")
			}
		}
	}()

	fmt.Printf("Run HTTP server on port:%v\n\r", httpPort)
	http.ListenAndServe(":"+httpPort, nil)
}
