package traqchat

import traqbot "github.com/traPtitech/traq-bot"

type Pattern struct {
	Func        ResFunc
	NeedMention bool
}

func (q *Pattern) CanExecute(payload *traqbot.MessageCreatedPayload, id string) bool {
	if payload.Message.User.Bot {
		return false
	}

	if q.NeedMention {
		for _, v := range payload.Message.Embedded {
			if v.Type == "user" && v.ID == id {
				return true
			}
		}

		return false
	}

	return true
}
