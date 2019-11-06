package listen

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"sync"
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

	var wg sync.WaitGroup

	// ONE GO ROUTINE ONLY
	wg.Add(1)
	go func() {
		defer wg.Done()

		pubSub := n.client.Subscribe(n.queue)
		defer pubSub.Close()
		// ONE FOR ONLY
		for {
			msg, _ := pubSub.ReceiveMessage()
			log.Printf("Recieved Message on Channel (%s): %s", msg.Channel, msg.Payload)
		}
	}()

	wg.Wait()
	return nil
}
