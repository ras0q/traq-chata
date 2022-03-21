package traqchat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/antihax/optional"
	"github.com/gofrs/uuid"
	traq "github.com/sapphi-red/go-traq"
	traqbot "github.com/traPtitech/traq-bot"
)

type (
	// Configuration of the bot
	TraqChat struct {
		ID                uuid.UUID // Bot uuid
		UserID            uuid.UUID // Bot user uuid
		AccessToken       string
		VerificationToken string
		Client            *traq.APIClient
		Auth              context.Context
		Writer            io.Writer
		Handlers          traqbot.EventHandlers
		Matchers          map[*regexp.Regexp]pattern
		Stamps            map[string]string
	}

	// Match pattern
	pattern struct {
		Func        ResponseFunc
		NeedMention bool
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

	stamps, _, err := client.StampApi.GetStamps(auth, &traq.StampApiGetStampsOpts{})
	if err != nil {
		log.Fatal(err)
	}

	stampsMap := make(map[string]string)
	for _, s := range stamps {
		stampsMap[s.Name] = s.Id
	}

	q := &TraqChat{
		ID:                uuid.FromStringOrNil(id),
		UserID:            uuid.FromStringOrNil(uid),
		AccessToken:       at,
		VerificationToken: vt,
		Client:            client,
		Auth:              auth,
		Writer:            os.Stdout,
		Handlers:          traqbot.EventHandlers{},
		Matchers:          map[*regexp.Regexp]pattern{},
		Stamps:            stampsMap,
	}

	q.Handlers.SetMessageCreatedHandler(func(payload *traqbot.MessageCreatedPayload) {
		for m, p := range q.Matchers {
			if m.MatchString(payload.Message.Text) && p.canExecute(payload, q.UserID) {
				if err := p.Func(&Response{
					tc:      *q,
					Payload: Payload{*payload},
				}); err != nil {
					fmt.Fprintln(q.Writer, err.Error())
				}
			}
		}
	})

	return q
}

func NewAndStart(id string, uid string, at string, vt string) {
	q := New(id, uid, at, vt)
	q.Start()
}

func (q *TraqChat) Start() {
	server := traqbot.NewBotServer(q.VerificationToken, q.Handlers)
	log.Fatal(server.ListenAndServe(":80"))
}

func (q *TraqChat) SetWriter(w io.Writer) {
	// Default writer is os.Stdout
	q.Writer = w
}

func (q *TraqChat) Hear(re *regexp.Regexp, f ResponseFunc) error {
	if _, ok := q.Matchers[re]; ok {
		return errors.New("Already Exists")
	}

	q.Matchers[re] = pattern{
		Func:        f,
		NeedMention: false,
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
	if _, ok := q.Matchers[re]; ok {
		return errors.New("Already Exists")
	}

	q.Matchers[re] = pattern{
		Func:        f,
		NeedMention: true,
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

func (r *Response) Send(content string) (traq.Message, error) {
	message, _, err := r.tc.Client.MessageApi.PostMessage(r.tc.Auth, r.Message.ChannelID, &traq.MessageApiPostMessageOpts{
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

func (r *Response) Reply(content string) (traq.Message, error) {
	message, _, err := r.tc.Client.MessageApi.PostMessage(r.tc.Auth, r.Message.ChannelID, &traq.MessageApiPostMessageOpts{
		PostMessageRequest: optional.NewInterface(traq.PostMessageRequest{
			Content: fmt.Sprintf("@%s %s", r.Message.User.Name, content),
			Embed:   true,
		}),
	})
	if err != nil {
		log.Println(fmt.Errorf("failed to reply a message: %w", err))

		return traq.Message{}, err
	}

	return message, nil
}

func (r *Response) AddStamp(stampName string) error {
	sid, ok := r.tc.Stamps[stampName]
	if !ok {
		return fmt.Errorf("stamp \"%s\" not found", stampName)
	}

	_, err := r.tc.Client.MessageApi.AddMessageStamp(r.tc.Auth, r.Message.ID, sid, &traq.MessageApiAddMessageStampOpts{})
	if err != nil {
		log.Println(fmt.Errorf("failed to add a stamp: %w", err))

		return err
	}

	return err
}

func (q *pattern) canExecute(payload *traqbot.MessageCreatedPayload, uid uuid.UUID) bool {
	if payload.Message.User.Bot {
		return false
	}

	if q.NeedMention {
		for _, v := range payload.Message.Embedded {
			if v.Type == "user" && v.ID == uid.String() {
				return true
			}
		}

		return false
	}

	return true
}
