package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/bylucasqueiroz/ddwrapper"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
)

func main() {
	// Configures DDWrapper
	dd := ddwrapper.New(ddwrapper.Config{
		ServiceName: "my-service",
		Environment: "prod",
		DDHost:      "localhost:8126",
		Enabled:     true,
	})
	defer dd.Stop()

	// 1. Instruments Gin
	r := gin.Default()
	dd.WrapGin(r)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 2. Instruments AWS SDK v2
	cfg, _ := config.LoadDefaultConfig(context.Background())
	dd.WrapAWSConfig(&cfg)
	// Use cfg with AWS services...

	// 3. Instruments Kafka
	producer, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	dd.WrapKafkaProducer(producer)
	defer producer.Close()
	// Use producer...

	consumer, _ := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "my-group",
	})
	dd.WrapKafkaConsumer(consumer)
	defer consumer.Close()
	// Use consumer...

	// 4. Instruments HTTP
	http.Handle("/hello", dd.WrapHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})))
	client := dd.WrapHTTPClient(&http.Client{})
	defer client.CloseIdleConnections()
	// Use client...

	// 5. Custom Metric
	dd.RecordCustomMetric(context.Background(), ddwrapper.CustomMetric{
		Name:  "myapp.request_duration",
		Tags:  []string{"endpoint:/ping"},
		Value: 150.5,
		Type:  "histogram",
	})

	// Starts the Gin server
	r.Run(":8080")
}
