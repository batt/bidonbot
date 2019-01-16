package slackbot

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"bufio"

	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

type SimpleAction func(*Bot, *slack.Msg)
type Action func(*Bot, *slack.Msg, ...string)

type Bot struct {
	Name   string
	UserID string

	config Config
	client *slack.Client
	rtm    *slack.RTM
	logger *logrus.Logger

	actions map[*regexp.Regexp]Action
	defact  SimpleAction
}

type Config struct {
	Offline bool
}

func New(token string, conf Config) *Bot {
	client := slack.New(token)
	logger := logrus.New()

	bot := &Bot{
		config:  conf,
		client:  client,
		rtm:     client.NewRTM(),
		logger:  logger,
		actions: make(map[*regexp.Regexp]Action),
	}

	return bot
}

func (bot *Bot) Start() error {
	if bot.config.Offline {
		return bot.startLocal()
	}

	return bot.startRTM()
}

func (bot *Bot) RespondTo(match string, action Action) {
	bot.actions[regexp.MustCompile(match)] = action
}

func (bot *Bot) DefaultResponse(action SimpleAction) {
	bot.defact = action
}

func (bot *Bot) Message(channel string, msg string) {
	if bot.config.Offline {
		fmt.Printf("< %s\n", msg)
	} else {
		bot.client.PostMessage(channel, msg, slack.NewPostMessageParameters())
	}
}

func (bot *Bot) startLocal() error {
	log := bot.logger
	br := bufio.NewReader(os.Stdin)

	log.Infoln("Running in local mode")

	for {
		fmt.Print("\033[50D")
		fmt.Print("\033[K")
		fmt.Print("> ")
		cmd, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		bot.handleMsg(&slack.Msg{
			Text: cmd,
		})
	}

	return nil
}

func (bot *Bot) startRTM() error {
	var filter filterer

	rtm := bot.rtm
	log := bot.logger

	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			bot.UserID = ev.Info.User.ID
			bot.Name = ev.Info.User.ID

			filter = newDirectFilter(bot.UserID)

			log.Infof("%s is online @ %s", bot.Name, ev.Info.Team.Name)
			log.Debugln("Bot info:", ev.Info)
			log.Debugln("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			if filter.filter(&ev.Msg) {
				log.Debugf("Message: %v\n", ev)
				bot.handleMsg(&ev.Msg)
			}

		case *slack.RTMError:
			log.Errorf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			return errors.New("Invalid credentials")

		default:
			// Can be used to handle custom events.
			// See: https://github.com/danackerson/bender-slackbot
		}
	}

	return nil
}

func (bot *Bot) handleMsg(msg *slack.Msg) {
	txt := bot.cleanupMsg(msg.Text)

	for match, action := range bot.actions {
		if matches := match.FindAllStringSubmatch(txt, -1); matches != nil {
			action(bot, msg, matches[0]...)
			return
		}
	}
	bot.defact(bot, msg)
}

func (bot *Bot) cleanupMsg(msg string) string {
	return strings.TrimLeft(strings.TrimSpace(msg), "<@"+bot.UserID+"> ")
}
