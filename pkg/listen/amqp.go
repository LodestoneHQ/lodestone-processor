package listen

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type AmqpListen struct {
	client   *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

func (n *AmqpListen) Init(config map[string]string) error {
	n.exchange = config["exchange"]
	n.queue = config["queue"]

	client, err := amqp.Dial(config["amqp-url"])
	if err != nil {
		return err
	}
	n.client = client

	//test connection
	ch, err := client.Channel()
	if err != nil {
		return err
	}
	n.channel = ch

	err = ch.ExchangeDeclare(
		n.exchange,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		n.queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	return err
}

func (n *AmqpListen) Subscribe(processor func(body []byte) error) error {
	fmt.Println("Subscribe to events..")

	err := n.channel.QueueBind(
		n.queue,    // queue name
		"",         // routing key
		n.exchange, // exchange
		false,
		nil)
	if err != nil {
		return err
	}

	msgs, err := n.channel.Consume(
		n.queue, // queue
		"",      // consumer
		true,    // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
			if err := processor(d.Body); err != nil {
				log.Printf("Error when processing document: %s", err)
				//TODO: add to the dead letter queue (for further processing later)
			}
		}
	}()

	log.Printf(" Waiting for logs. To exit press CTRL+C")
	<-forever

	return nil
}

func (n *AmqpListen) Close() error {
	if err := n.client.Close(); err != nil {
		return err
	}
	if err := n.channel.Close(); err != nil {
		return err
	}
	return nil
}
