package ddwrapper

import (
	"context"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// WrapKafkaProducer adds tracing to the Kafka Producer without changing its type
func (w *DDWrapper) WrapKafkaProducer(producer *kafka.Producer) *kafka.Producer {
	if !w.config.Enabled {
		return producer
	}

	// Intercepts producer events to add tracing
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				// Adds tracing when the message is sent
				if ev.TopicPartition.Error == nil {
					span, _ := tracer.StartSpanFromContext(context.Background(), "kafka.produce",
						tracer.ServiceName(w.config.ServiceName),
						tracer.ResourceName(*ev.TopicPartition.Topic),
						tracer.SpanType("kafka"),
						tracer.Tag("partition", ev.TopicPartition.Partition),
						tracer.Tag("offset", ev.TopicPartition.Offset),
					)
					span.Finish()
				}
			}
		}
	}()

	return producer
}

// WrapKafkaConsumer adds tracing to the Kafka Consumer without changing its type
func (w *DDWrapper) WrapKafkaConsumer(consumer *kafka.Consumer) *kafka.Consumer {
	if !w.config.Enabled {
		return consumer
	}

	// No direct interception is needed here; tracing will be added during consumption
	// The user should call TrackKafkaMessage in the consumption loop
	return consumer
}

// TrackKafkaMessage adds tracing to a consumed message
func (w *DDWrapper) TrackKafkaMessage(msg *kafka.Message) {
	if !w.config.Enabled {
		return
	}

	span, _ := tracer.StartSpanFromContext(context.Background(), "kafka.consume",
		tracer.ServiceName(w.config.ServiceName),
		tracer.ResourceName(*msg.TopicPartition.Topic),
		tracer.SpanType("kafka"),
		tracer.Tag("partition", msg.TopicPartition.Partition),
		tracer.Tag("offset", msg.TopicPartition.Offset),
	)
	span.Finish()
}
