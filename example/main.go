package main

import (
	"fmt"
	"os"
	"regexp"

	traqchat "github.com/Ras96/traq-chat"
	"github.com/gofrs/uuid"
)

func main() {
	q := traqchat.New(
		uuid.FromStringOrNil(os.Getenv("BOT_ID")),
		uuid.FromStringOrNil(os.Getenv("BOT_USER_ID")),
		os.Getenv("BOT_ACCESS_TOKEN"),
		os.Getenv("BOTVERIFICATION_TOKEN"),
	)

	q.Hear(regexp.MustCompile(`ping`), func(res *traqchat.Response) error {
		res.Send("pong!")

		return nil
	})

	q.Respond(regexp.MustCompile(`Hello`), hello)

	q.Start()
}

func hello(res *traqchat.Response) error {
	res.Reply(fmt.Sprintf("Hello, %s\n", res.Message.User.DisplayName))

	return nil
}
