package ekafka

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/segmentio/kafka-go"
)

func EnsureTopic(broker string, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial(
		"tcp",
		net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)),
	)
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     3,
		ReplicationFactor: 1,
	})

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Println("topic already exists:", topic)
			return nil
		}
		return err
	}

	log.Println("topic created:", topic)
	return nil
}
