package slackbot

import (
	"strings"

	"github.com/nlopes/slack"
)

type filterer interface {
	filter(msg *slack.Msg) bool
}

type multiFilter struct {
	filters []filterer
}

type directFilter struct {
	id string
}

func newDirectFilter(id string) *directFilter {
	return &directFilter{
		id: id,
	}
}

func (f *directFilter) filter(msg *slack.Msg) bool {
	return msg.Type == "message" &&
		(msg.SubType != "message_deleted" && msg.SubType != "bot_message") &&
		msg.User != f.id &&
		(strings.HasPrefix(msg.Text, "<@"+f.id+">") || strings.HasPrefix(msg.Channel, "D"))
}

func newMultiFilter(filters ...filterer) *multiFilter {
	return &multiFilter{
		filters: filters,
	}
}

func (f *multiFilter) filter(msg *slack.Msg) bool {
	for _, filter := range f.filters {
		if !filter.filter(msg) {
			return false
		}
	}
	return true
}
