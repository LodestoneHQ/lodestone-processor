package listen

import (
	"fmt"
	"github.com/go-redis/redis"
)

type RedisListen struct {
	client *redis.Client
	queue  string
}

func (n *RedisListen) Init(config map[string]string) error {
	n.client = redis.NewClient(&redis.Options{
		Addr:     config["addr"],     //"localhost:6379",
		Password: config["password"], //"", // no password set
		DB:       0,                  // use default DB
	})
	n.queue = config["queue"]

	pong, err := n.client.Ping().Result()
	fmt.Println(pong, err)

	return err
}

func (n *RedisListen) Subscribe() error {
	fmt.Println("Subscribe to events on model")

	// ONE GO ROUTINE ONLY
	go func() {
		pubSub := n.client.Subscribe(
			n.queue,
		)
		// ONE FOR ONLY
		for {
			msg, _ := pubSub.ReceiveMessage()
			switch msg.Channel {
			case n.queue:
				go func() {
					// DO SOMETHING WITH msg
				}()
			case "channel2":
				go func() {
					// DO SOMETHING WITH msg
				}()
				// MORE 7 CASES
			}
		}
	}()
	return nil
}
