package ekafka

import (
	"fmt"
	"log"
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

	log.Println("controller:", controller.Host, controller.Port)

	host := controller.Host
	port := controller.Port

	if host == "" || port == 0 {
		log.Println("fallback to broker")
		host = "kafka"
		port = 9092
	}

	controllerConn, err := kafka.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", host, port),
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
