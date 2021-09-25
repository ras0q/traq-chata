package traqchat

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/antihax/optional"
	traq "github.com/sapphi-red/go-traq"
	traqbot "github.com/traPtitech/traq-bot"
)

type TraqChat struct {
	ID                string // Bot uuid
	UserID            string // Bot user uuid
	AccessToken       string
	VerificationToken string
	Client            *traq.APIClient
	Auth              context.Context
	Handlers          traqbot.EventHandlers
	Matchers          map[*regexp.Regexp]Pattern
}

type Payload struct {
	traqbot.MessageCreatedPayload
}

type Res struct {
	TraqChat
	Payload
}

type ResFunc = func(*Res) error

func newPayload(p traqbot.MessageCreatedPayload) Payload {
	return Payload{p}
}

func newRes(c TraqChat, p Payload) *Res {
	return &Res{
		TraqChat: c,
		Payload:  p,
	}
}

func New(id, uid, at, vt string) *TraqChat {
	client := traq.NewAPIClient(traq.NewConfiguration())
	auth := context.WithValue(context.Background(), traq.ContextAccessToken, at)

	q := &TraqChat{
		ID:                id,
		UserID:            uid,
		AccessToken:       at,
		VerificationToken: vt,
		Client:            client,
		Auth:              auth,
		Handlers:          traqbot.EventHandlers{},
		Matchers:          map[*regexp.Regexp]Pattern{},
	}

	q.Handlers.SetMessageCreatedHandler(func(payload *traqbot.MessageCreatedPayload) {
		for m, p := range q.Matchers {
			if m.MatchString(payload.Message.Text) && p.CanExecute(payload, q.UserID) {
				p.Func(newRes(*q, newPayload(*payload)))
			}
		}
	})

	return q
}

func (q *TraqChat) Hear(re *regexp.Regexp, f ResFunc) error {
	if _, ok := q.Matchers[re]; ok {
		return errors.New("Already Exists")
	}

	q.Matchers[re] = Pattern{
		Func:        f,
		NeedMention: false,
	}

	return nil
}

func (q *TraqChat) Respond(re *regexp.Regexp, f ResFunc) error {
	if _, ok := q.Matchers[re]; ok {
		return errors.New("Already Exists")
	}

	q.Matchers[re] = Pattern{
		Func:        f,
		NeedMention: true,
	}

	return nil
}

func (q *TraqChat) Start() {
	server := traqbot.NewBotServer(q.VerificationToken, q.Handlers)
	log.Fatal(server.ListenAndServe(":80"))
}

func (r *Res) Send(content string) (traq.Message, error) {
	message, _, err := r.Client.MessageApi.PostMessage(r.Auth, r.Message.ChannelID, &traq.MessageApiPostMessageOpts{
		PostMessageRequest: optional.NewInterface(traq.PostMessageRequest{
			Content: content,
			Embed:   true,
		}),
	})

	if err != nil {
		log.Println(fmt.Errorf("failed to send a message: %w", err))

		return traq.Message{}, err
	}

	return message, nil
}

func (r *Res) Reply(content string) (traq.Message, error) {
	reply := fmt.Sprintf("@%s %s", r.Message.User.Name, content)
	message, _, err := r.Client.MessageApi.PostMessage(r.Auth, r.Message.ChannelID, &traq.MessageApiPostMessageOpts{
		PostMessageRequest: optional.NewInterface(traq.PostMessageRequest{
			Content: reply,
			Embed:   true,
		}),
	})

	if err != nil {
		log.Println(fmt.Errorf("failed to reply a message: %w", err))

		return traq.Message{}, err
	}

	return message, nil
}
