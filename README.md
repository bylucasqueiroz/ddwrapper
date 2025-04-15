# ddwrapper - Instrumentation Library for Datadog in Go

`ddwrapper` is a Go library that simplifies integration with Datadog for tracing and metrics instrumentation in applications using Gin, Kafka (Confluent), AWS SDK v2, HTTP, and custom metrics. It provides a unified API to inject Datadog tracing middlewares and send custom metrics, reducing the complexity of configuring each library individually.

## Features

### Automatic Instrumentation:
- **Gin**: Adds tracing for HTTP routes.
- **Kafka (Confluent)**: Monitors sent and received messages.
- **AWS SDK v2**: Instruments AWS API calls.
- **HTTP**: Adds tracing for HTTP servers and clients.
- **Custom Metrics**: Allows sending custom metrics (gauge, count, histogram) to Datadog via DogStatsD.

### Centralized Configuration:
- Global control to enable/disable tracing and define service/environment.

### Extensibility:
- Easy to add support for new libraries.

---

## Prerequisites

- **Go**: Version 1.18 or higher.
- **Datadog Agent**: Running locally (`localhost:8126` for traces, `localhost:8125` for metrics) or configured remotely.

### Dependencies:
```go
require (
    github.com/aws/aws-sdk-go-v2 v1.21.0
    github.com/confluentinc/confluent-kafka-go v1.9.2
    github.com/gin-gonic/gin v1.8.1
    gopkg.in/DataDog/dd-trace-go.v1 v1.48.0
)
```

---

## Installation

Add the library to your project:
```bash
go get github.com/your-org/ddwrapper
```

Ensure the Datadog agent is configured and running.

---

## Configuration

Create an instance of `DDWrapper` with the desired settings:
```go
package main

import (
    "github.com/your-org/ddwrapper"
)

func main() {
    dd := ddwrapper.New(ddwrapper.Config{
        ServiceName: "my-service",
        Environment: "prod",
        DDHost:      "localhost:8126",
        Enabled:     true,
    })
    defer dd.Stop()
}
```

### Configuration Options:
- **ServiceName**: Service name in Datadog (e.g., `my-service`).
- **Environment**: Application environment (e.g., `prod`, `staging`).
- **DDHost**: Datadog agent address (default: `localhost:8126`).
- **Enabled**: Enables (`true`) or disables (`false`) tracing.

---

## Implementations

### 1. Gin
Adds automatic tracing for HTTP routes in the Gin framework.

#### How it works:
- Applies the `gindd.Middleware` from `dd-trace-go` to the Gin router.
- Each request generates a span in Datadog with details like method, URL, and status.

#### Example:
```go
import "github.com/gin-gonic/gin"

r := gin.Default()
dd.WrapGin(r)
r.GET("/ping", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "pong"})
})
r.Run(":8080")
```

#### Output in Datadog:
- Spans with type `web`, resource name like `/ping`, and tags like `http.method=GET`.

---

### 2. Kafka (Confluent)
Instruments Confluent Kafka producers and consumers to monitor sent and received messages.

#### How it works:
- For producers, monitors the `Events()` channel and creates spans for successfully sent messages.
- For consumers, requires calling `TrackKafkaMessage` for each consumed message, creating corresponding spans.
- Spans include tags like `topic`, `partition`, and `offset`.

#### Example:
```go
import "github.com/confluentinc/confluent-kafka-go/kafka"

// Producer
producer, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
producer = dd.WrapKafkaProducer(producer)
defer producer.Close()
producer.Produce(&kafka.Message{
    TopicPartition: kafka.TopicPartition{Topic: stringp("my-topic"), Partition: kafka.PartitionAny},
    Value:          []byte("Hello, Kafka!"),
}, nil)

// Consumer
consumer, _ := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers": "localhost:9092",
    "group.id":          "my-group",
    "auto.offset.reset": "earliest",
})
consumer = dd.WrapKafkaConsumer(consumer)
defer consumer.Close()
consumer.SubscribeTopics([]string{"my-topic"}, nil)
for {
    msg, err := consumer.ReadMessage(-1)
    if err != nil {
        continue
    }
    dd.TrackKafkaMessage(msg)
}

// Helper
func stringp(s string) *string {
    return &s
}
```

#### Output in Datadog:
- Spans `kafka.produce` and `kafka.consume` with tags like `topic=my-topic`, `partition`, `offset`.

---

### 3. AWS SDK v2
Instruments AWS SDK v2 API calls.

#### How it works:
- Adds a custom middleware to `aws.Config` to create spans for each API call.
- Extracts metadata like service (e.g., `s3`, `dynamodb`) and operation (e.g., `GetObject`).

#### Example:
```go
import (
    "context"
    "github.com/aws/aws-sdk-go-v2/config"
)

cfg, _ := config.LoadDefaultConfig(context.Background())
dd.WrapAWSConfig(&cfg)
// Use cfg with AWS services (e.g., S3, DynamoDB)
```

#### Output in Datadog:
- Spans like `aws.s3.request`, `aws.dynamodb.request` with tags `aws.service`, `aws.operation`.

---

### 4. HTTP
Adds tracing for standard HTTP servers and clients.

#### How it works:
- For handlers, manually creates spans for each request with tags like `http.method` and `http.url`.
- For clients, uses `httpdd.WrapClient` from `dd-trace-go` to instrument HTTP requests.

#### Example:
```go
import "net/http"

// Handler
http.Handle("/hello", dd.WrapHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
})))
http.ListenAndServe(":8080", nil)

// Client
client := dd.WrapHTTPClient(&http.Client{})
req, _ := http.NewRequest("GET", "http://example.com", nil)
client.Do(req)
```

#### Output in Datadog:
- Spans `http.request` with tags like `http.method=GET`, `http.url`.

---

### 5. Custom Metrics
Allows sending custom metrics (gauge, count, histogram) to Datadog via DogStatsD.

#### How it works:
- Uses the DogStatsD client to send metrics with name, value, type, and tags.
- Supports global tags (e.g., `env=prod`) and user-specific tags.

#### Example:
```go
import "context"

dd.RecordCustomMetric(context.Background(), ddwrapper.CustomMetric{
    Name:  "myapp.request_duration",
    Tags:  []string{"endpoint:/ping"},
    Value: 150.5,
    Type:  "histogram",
})
```

#### Output in Datadog:
- Metrics like `myapp.request_duration` with tags `endpoint:/ping`, `env=prod`.

---

## Full Example
```go
package main

import (
    "context"
    "net/http"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/confluentinc/confluent-kafka-go/kafka"
    "github.com/gin-gonic/gin"
    "github.com/your-org/ddwrapper"
)

func main() {
    // Configure DDWrapper
    dd := ddwrapper.New(ddwrapper.Config{
        ServiceName: "my-service",
        Environment: "prod",
        DDHost:      "localhost:8126",
        Enabled:     true,
    })
    defer dd.Stop()

    // Gin
    r := gin.Default()
    dd.WrapGin(r)
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    // Kafka Producer
    producer, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
    producer = dd.WrapKafkaProducer(producer)
    producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{Topic: stringp("my-topic"), Partition: kafka.PartitionAny},
        Value:          []byte("Hello, Kafka!"),
    }, nil)

    // Kafka Consumer
    consumer, _ := kafka.NewConsumer(&kafka.ConfigMap{
        "bootstrap.servers": "localhost:9092",
        "group.id":          "my-group",
        "auto.offset.reset": "earliest",
    })
    consumer = dd.WrapKafkaConsumer(consumer)
    go func() {
        consumer.SubscribeTopics([]string{"my-topic"}, nil)
        for {
            msg, err := consumer.ReadMessage(-1)
            if err != nil {
                continue
            }
            dd.TrackKafkaMessage(msg)
        }
    }()

    // AWS SDK v2
    cfg, _ := config.LoadDefaultConfig(context.Background())
    dd.WrapAWSConfig(&cfg)

    // HTTP
    http.Handle("/hello", dd.WrapHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })))
    client := dd.WrapHTTPClient(&http.Client{})
    req, _ := http.NewRequest("GET", "http://example.com", nil)
    client.Do(req)

    // Custom Metric
    dd.RecordCustomMetric(context.Background(), ddwrapper.CustomMetric{
        Name:  "myapp.request_duration",
        Tags:  []string{"endpoint:/ping"},
        Value: 150.5,
        Type:  "histogram",
    })

    // Start server
    r.Run(":8080")
}

func stringp(s string) *string {
    return &s
}
```

---

## Notes

- **Performance**: Instrumentation adds minimal overhead, but test in staging environments before production use.
- **Datadog Agent**: Ensure the agent is correctly configured to receive traces and metrics.
- **Extensibility**: To add support for new libraries, create similar `Wrap*` methods using `dd-trace-go` contrib packages or manual tracing.

---

## Troubleshooting

- **Compilation Errors**: Check dependencies in `go.mod` and clean the cache (`go clean -modcache; go mod tidy`).
- **Missing Spans/Metrics**: Verify `DDHost` and ensure the Datadog agent is running.
- **Kafka Consumer**: Remember to call `TrackKafkaMessage` for each consumed message.

---

## Contribution

Contributions are welcome! Open an issue or pull request at [github.com/bylucasqueiroz/ddwrapper](https://github.com/bylucasqueiroz/ddwrapper).

---

## License

MIT License