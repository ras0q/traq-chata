# traq-chat

Make creating traQBot easier.

## Example

```go
package main

import (
	"os"

	traqchat "github.com/Ras96/traq-chat"
)

func main() {
		// Note: メンション埋め込みは自動でオンになります
	q := traqchat.New(os.Getenv("BOT_ID"), os.Getenv("ACCESS_TOKEN"), os.Getenv("VERIFICATION_TOKEN"))

	q.Hear(`ping`, func(payload *traqchat.Payload) {
		traqchat.Reply(q, payload, "pong!")
	})

	q.Start()
}

```
