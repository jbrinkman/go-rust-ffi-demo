package main

import (
	"fmt"
	"time"

	"github.com/jbrinkman/go-rust-ffi/go/pubsub"
)

func main() {
	// Example of using the pubsub system
	fmt.Println("PubSub Example")

	// Subscribe with a callback
	err := pubsub.Subscribe("subscriber1", "news", func(topic, message string) {
		fmt.Printf("Callback received: Topic=%s, Message=%s\n", topic, message)
	})
	if err != nil {
		fmt.Printf("Error subscribing: %v\n", err)
		return
	}

	// Subscribe without a callback (using message queue)
	err = pubsub.Subscribe("subscriber2", "news", nil)
	if err != nil {
		fmt.Printf("Error subscribing: %v\n", err)
		return
	}

	// Publish some messages
	fmt.Println("Publishing messages...")
	pubsub.Publish("news", "Breaking news: Go-Rust FFI works!")
	pubsub.Publish("news", "More news: Pub/Sub system is operational")

	// Give some time for callbacks to execute
	time.Sleep(100 * time.Millisecond)

	// Check and retrieve messages for subscriber2
	fmt.Println("\nChecking messages for subscriber2:")
	if pubsub.HasMessages("subscriber2", "") {
		fmt.Println("Subscriber2 has messages")

		// Get all available messages
		for pubsub.HasMessages("subscriber2", "") {
			msg, err := pubsub.GetMessage("subscriber2", "")
			if err != nil {
				fmt.Printf("Error getting message: %v\n", err)
				break
			}
			fmt.Printf("Retrieved message: Topic=%s, Content=%s\n", msg.Topic, msg.Content)
		}
	} else {
		fmt.Println("No messages for subscriber2")
	}

	// Unsubscribe from specific topic
	fmt.Println("\nUnsubscribing subscriber1 from news topic")
	err = pubsub.Unsubscribe("subscriber1", "news")
	if err != nil {
		fmt.Printf("Error unsubscribing: %v\n", err)
	}

	// Publish another message (subscriber1 won't receive it)
	pubsub.Publish("news", "Final update: Subscriber1 won't see this")

	// Give some time for callbacks to execute
	time.Sleep(100 * time.Millisecond)

	// Check messages for subscriber2 again
	fmt.Println("\nChecking messages for subscriber2 again:")
	if pubsub.HasMessages("subscriber2", "") {
		msg, err := pubsub.GetMessage("subscriber2", "")
		if err != nil {
			fmt.Printf("Error getting message: %v\n", err)
		} else {
			fmt.Printf("Retrieved message: Topic=%s, Content=%s\n", msg.Topic, msg.Content)
		}
	} else {
		fmt.Println("No messages for subscriber2")
	}

	// Unsubscribe from all topics
	fmt.Println("\nUnsubscribing subscriber2 from all topics")
	err = pubsub.Unsubscribe("subscriber2", "")
	if err != nil {
		fmt.Printf("Error unsubscribing: %v\n", err)
	}
}
