# Skadi Go SDK
[hack-fan/skadi](https://github.com/hack-fan/skadi) is a cloud message center,
you can send your message/job/command to it by several ways, Slack/Teams/Wechat etc...
then your agent will get it, do anything you defined, return the result.

This is golang sdk for hack-fan/skadi, pull your jobs from skadi server.

## Example
Prepare your `TOKEN` first.

There is several example agent by our team:
* https://github.com/hack-fan/skadi-agent-shell can run shell commands.
* https://github.com/hack-fan/skadi-agent-docker can restart docker swarm service.

You can only start one worker per TOKEN.
```go
package main

import (
    "context"
    "fmt"
    "os/signal"
    "syscall"

    "github.com/hack-fan/skadigo"
)

func handler(id,msg string) (string, error) {
    fmt.Printf("received command message: %s %s", id, msg)
    return msg,nil
}

func main() {
    // system signals - for graceful shutdown or restart
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // skadi agent
    agent := skadigo.New("YOUR-TOKEN", "https://api.letserver.run", nil)
    agent.Start(ctx, handler, 0)
}
```

You can use the agent for sending messages anywhere.

```go
    agent.SendInfo("Hello World")
```
