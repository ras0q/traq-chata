package traqchat

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/antihax/optional"
	"github.com/jinzhu/copier"
	traq "github.com/sapphi-red/go-traq"
	traqbot "github.com/traPtitech/traq-bot"
)

type TraqChat struct {
	ID                string // Bot uuid
	AccessToken       string
	VerificationToken string
	Client            *traq.APIClient
	Auth              context.Context
	Handlers          traqbot.EventHandlers
	Matchers          map[*regexp.Regexp]Pattern
	Embed             bool
}

type Payload traqbot.MessageCreatedPayload

func New(id, at, vt string, embed bool) *TraqChat {
	client := traq.NewAPIClient(traq.NewConfiguration())
	auth := context.WithValue(context.Background(), traq.ContextAccessToken, at)

	q := &TraqChat{
		ID:                id,
		AccessToken:       at,
		VerificationToken: vt,
		Client:            client,
		Auth:              auth,
		Handlers:          traqbot.EventHandlers{},
		Embed:             embed,
	}

	q.Handlers.SetMessageCreatedHandler(func(payload *traqbot.MessageCreatedPayload) {
		for m, p := range q.Matchers {
			if m.MatchString(payload.Message.Text) && p.CanExecute(payload, q.ID) {
				pl := Payload{}

				copier.Copy(&pl, &payload)

				p.Func(&pl)
			}
		}
	})

	return q
}

func (q *TraqChat) Hear(restr string, f func(*Payload)) error {
	re, err := regexp.Compile(restr)
	if err != nil {
		return err
	}

	if _, ok := q.Matchers[re]; ok {
		return errors.New("Already Exists")
	}

	q.Matchers[re] = Pattern{
		Func:        f,
		NeedMention: false,
	}

	return nil
}

func (q *TraqChat) Respond(restr string, f func(*traqbot.MessageCreatedPayload)) error {
	re, err := regexp.Compile(restr)
	if err != nil {
		return err
	}

	if _, ok := q.Matchers[re]; ok {
		return errors.New("Already Exists")
	}

	q.Matchers[re] = Pattern{
		Func:        f,
		NeedMention: true,
	}

	return nil
}

func (q *TraqChat) Send(payload *Payload, content string) (traq.Message, error) {
	message, _, err := q.Client.MessageApi.PostMessage(q.Auth, payload.Message.ChannelID, &traq.MessageApiPostMessageOpts{
		PostMessageRequest: optional.NewInterface(traq.PostMessageRequest{
			Content: content,
			Embed:   q.Embed,
		}),
	})

	if err != nil {
		log.Println(fmt.Errorf("failed to send a message: %w", err))

		return traq.Message{}, err
	}

	return message, nil
}

func (q *TraqChat) Reply(payload *Payload, content string) (traq.Message, error) {
	reply := fmt.Sprintf(
		"!{\"type\":\"user\",\"raw\":\"@%s\",\"id\":\"%s\"}\n%s",
		payload.Message.User.Name,
		payload.Message.User.ID,
		content,
	)
	message, _, err := q.Client.MessageApi.PostMessage(q.Auth, payload.Message.ChannelID, &traq.MessageApiPostMessageOpts{
		PostMessageRequest: optional.NewInterface(traq.PostMessageRequest{
			Content: reply,
			Embed:   q.Embed,
		}),
	})

	if err != nil {
		log.Println(fmt.Errorf("failed to reply a message: %w", err))

		return traq.Message{}, err
	}

	return message, nil
}

func (q *TraqChat) Start() {
	server := traqbot.NewBotServer(q.VerificationToken, q.Handlers)
	log.Fatal(server.ListenAndServe(":80"))
}
