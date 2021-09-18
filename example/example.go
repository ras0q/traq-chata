package main

import (
	"os"

	traqchat "github.com/Ras96/traq-chat"
)

func main() {
	q := traqchat.New(os.Getenv("BOT_ID"), os.Getenv("ACCESS_TOKEN"), os.Getenv("VERIFICATION_TOKEN"))

	q.Hear(`ping`, func(q *traqchat.TraqChat, payload *traqchat.Payload) {
		traqchat.Send(q, payload, "pong!")
	})

	q.Start()
}
