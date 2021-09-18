# traq-chat

Make creating traQBot more easy.

## Example

```go
package main

import (
	"os"

	traqchat "github.com/Ras96/traq-chat"
)

func main() {
	q := traqchat.New(os.Getenv("BOT_ID"), os.Getenv("ACCESS_TOKEN"), os.Getenv("VERIFICATION_TOKEN"), true)

	q.Hear(`ping`, func(payload *traqchat.Payload) {
		q.Reply(payload, "pong!")
	})

	q.Start()
}

```
