package listen

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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

	//define the deadletter queue
	err = ch.ExchangeDeclare(
		"errors",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(
		"errors", // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}
	err = n.channel.QueueBind(
		"errors", // queue name
		"",       // routing key
		"errors", // exchange
		false,
		nil)
	if err != nil {
		return err
	}

	//define active exchange & queue

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

	//setup dead-letter queu
	args := make(amqp.Table)
	args["x-dead-letter-exchange"] = "errors"
	args["x-dead-letter-routing-key"] = n.queue

	_, err = ch.QueueDeclare(
		n.queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		args,    // arguments
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
		false,   // auto-ack
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

				//add to the dead letter queue (for further processing later)
				if err := d.Reject(false); err != nil {
					log.Printf("Error while adding document to dead-letter-queue")
				}
			} else {
				if err := d.Ack(false); err != nil {
					log.Printf("Error while notifying successful processing")
				}
			}
		}
	}()

	log.Printf("Waiting for logs. To exit press CTRL+C")
	<-forever

	log.Printf("ERROR: Should never get here. %s", err)
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
