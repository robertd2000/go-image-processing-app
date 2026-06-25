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

	host := controller.Host
	port := controller.Port

	if host == "" || port == 0 {
		// fallback: use the original broker address
		host, port = splitBroker(broker)
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

func splitBroker(addr string) (string, int) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		return parts[0], 9092
	}
	return "localhost", 9092
}
