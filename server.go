package traqchat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	traq "github.com/traPtitech/go-traq"
	traqbot "github.com/traPtitech/traq-bot"
)

var embedTrue = true

type (
	// Configuration of the bot
	TraqChat struct {
		id                string // Bot uuid
		userID            string // Bot user uuid
		accessToken       string
		verificationToken string
		writer            io.Writer
		handlers          traqbot.EventHandlers
		matchers          map[*regexp.Regexp]pattern
		stamps            map[string]string

		// Wrapper for traPtitech/go-traq
		TraqAPIClient *traq.APIClient
		TraqAPIAuth   context.Context
	}

	// Match pattern
	pattern struct {
		responseFunc ResponseFunc
		needMention  bool
	}

	// Response Information
	Response struct {
		tc TraqChat
		Payload
	}

	// Wrapper for traqbot.MessageCreatedPayload
	Payload struct {
		traqbot.MessageCreatedPayload
	}

	// Response function
	ResponseFunc func(*Response) error

	// Response function
	MustResponseFunc func(*Response)
)

func New(id string, uid string, at string, vt string) *TraqChat {
	client := traq.NewAPIClient(traq.NewConfiguration())
	auth := context.WithValue(context.Background(), traq.ContextAccessToken, at)

	stamps, _, err := client.StampApi.
		GetStamps(auth).
		Execute()
	if err != nil {
		log.Fatal(err)
	}

	stampsMap := make(map[string]string)
	for _, s := range stamps {
		stampsMap[s.Name] = s.Id
	}

	q := &TraqChat{
		id:                id,
		userID:            uid,
		accessToken:       at,
		verificationToken: vt,
		writer:            os.Stdout,
		handlers:          traqbot.EventHandlers{},
		matchers:          map[*regexp.Regexp]pattern{},
		stamps:            stampsMap,
		TraqAPIClient:     client,
		TraqAPIAuth:       auth,
	}

	q.handlers.SetMessageCreatedHandler(func(payload *traqbot.MessageCreatedPayload) {
		for m, p := range q.matchers {
			if m.MatchString(payload.Message.Text) && p.canExecute(payload, q.userID) {
				if err := p.responseFunc(&Response{
					tc:      *q,
					Payload: Payload{*payload},
				}); err != nil {
					fmt.Fprintln(q.writer, err.Error())
				}
			}
		}
	})

	return q
}

func NewAndStart(id string, uid string, at string, vt string, port int) {
	q := New(id, uid, at, vt)
	q.Start(port)
}

func (q *TraqChat) Start(port int) {
	server := traqbot.NewBotServer(q.verificationToken, q.handlers)
	log.Fatal(server.ListenAndServe(fmt.Sprintf(":%d", port)))
}

func (q *TraqChat) SetWriter(w io.Writer) {
	// Default writer is os.Stdout
	q.writer = w
}

func (q *TraqChat) Hear(re *regexp.Regexp, f ResponseFunc) error {
	if _, ok := q.matchers[re]; ok {
		return errors.New("Already Exists")
	}

	q.matchers[re] = pattern{
		responseFunc: f,
		needMention:  false,
	}

	return nil
}

func (q *TraqChat) HearF(re *regexp.Regexp, f MustResponseFunc) error {
	if err := q.Hear(re, func(r *Response) error {
		f(r)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (q *TraqChat) Respond(re *regexp.Regexp, f ResponseFunc) error {
	if _, ok := q.matchers[re]; ok {
		return errors.New("Already Exists")
	}

	q.matchers[re] = pattern{
		responseFunc: f,
		needMention:  true,
	}

	return nil
}

func (q *TraqChat) RespondF(re *regexp.Regexp, f MustResponseFunc) error {
	if err := q.Respond(re, func(r *Response) error {
		f(r)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *Response) Send(content string) (*traq.Message, error) {
	message, _, err := r.tc.TraqAPIClient.MessageApi.
		PostMessage(r.tc.TraqAPIAuth, r.Message.ChannelID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: content,
			Embed:   &embedTrue,
		}).
		Execute()
	if err != nil {
		log.Println(fmt.Errorf("failed to send a message: %w", err))

		return nil, err
	}

	return message, nil
}

func (r *Response) Reply(content string) (*traq.Message, error) {
	message, _, err := r.tc.TraqAPIClient.MessageApi.
		PostMessage(r.tc.TraqAPIAuth, r.Message.ChannelID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: fmt.Sprintf("@%s %s", r.Message.User.Name, content),
			Embed:   &embedTrue,
		}).
		Execute()
	if err != nil {
		log.Println(fmt.Errorf("failed to reply a message: %w", err))

		return nil, err
	}

	return message, nil
}

func (r *Response) AddStamp(stampName string) error {
	sid, ok := r.tc.stamps[stampName]
	if !ok {
		return fmt.Errorf("stamp \"%s\" not found", stampName)
	}

	_, err := r.tc.TraqAPIClient.MessageApi.
		AddMessageStamp(r.tc.TraqAPIAuth, r.Message.ID, sid).
		Execute()
	if err != nil {
		log.Println(fmt.Errorf("failed to add a stamp: %w", err))

		return err
	}

	return err
}

func (q *pattern) canExecute(payload *traqbot.MessageCreatedPayload, uid string) bool {
	if payload.Message.User.Bot {
		return false
	}

	if q.needMention {
		for _, v := range payload.Message.Embedded {
			if v.Type == "user" && v.ID == uid {
				return true
			}
		}

		return false
	}

	return true
}
