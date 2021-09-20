# traq-chat

Make creating traQBot easier.

## Example

```go
package main

import (
	"fmt"
	"os"

	traqchat "github.com/Ras96/traq-chat"
)

func main() {
	q := traqchat.New(
		os.Getenv("BOT_ID"),
		os.Getenv("BOT_USER_ID"),
		os.Getenv("BOT_ACCESS_TOKEN"),
		os.Getenv("BOTVERIFICATION_TOKEN"),
	)

	q.Hear(`ping`, func(res *traqchat.Res) {
		res.Send("pong!")
	})

	q.Respond(`Hello`, hello)

	q.Start()
}

func hello(res *traqchat.Res) {
	res.Reply(fmt.Sprintf("Hello, %s\n", res.Message.User.DisplayName))
}
```

## More Information for traQ Bot

- [traPtitech/traq-bot: traQ BOT用Goライブラリ](https://github.com/traPtitech/traq-bot)
