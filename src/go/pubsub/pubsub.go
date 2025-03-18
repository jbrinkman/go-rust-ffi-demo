package pubsub

// #cgo LDFLAGS: -L../../target/release -lpubsub_core
// #include <stdlib.h>
// #include <stdbool.h>
//
// typedef void (*message_callback)(const char* topic, const char* message, void* user_data);
//
// extern bool subscribe(const char* subscriber_id, const char* topic, message_callback callback, void* user_data);
// extern bool unsubscribe(const char* subscriber_id, const char* topic);
// extern bool publish(const char* topic, const char* message);
// extern bool get_next_message(const char* subscriber_id, const char* topic, char* out_topic, size_t out_topic_size, char* out_message, size_t out_message_size);
// extern bool has_messages(const char* subscriber_id, const char* topic);
//
// // Gateway function for the callback
// void callbackGateway(const char* topic, const char* message, void* user_data);
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// Maximum buffer size for messages
const (
	MaxTopicSize    = 256
	MaxMessageSize  = 4096
)

// MessageCallback is the Go type for message callbacks
type MessageCallback func(topic, message string)

// callbackRegistry keeps track of Go callbacks by subscriber ID
var callbackRegistry = struct {
	sync.RWMutex
	callbacks map[string]MessageCallback
}{
	callbacks: make(map[string]MessageCallback),
}

//export callbackGateway
func callbackGateway(topic *C.char, message *C.char, userData unsafe.Pointer) {
	subscriberID := C.GoString((*C.char)(userData))
	
	callbackRegistry.RLock()
	callback, exists := callbackRegistry.callbacks[subscriberID]
	callbackRegistry.RUnlock()
	
	if exists {
		callback(C.GoString(topic), C.GoString(message))
	}
}

// Subscribe registers a subscription to a topic with an optional callback
func Subscribe(subscriberID, topic string, callback MessageCallback) error {
	cSubscriberID := C.CString(subscriberID)
	defer C.free(unsafe.Pointer(cSubscriberID))
	
	cTopic := C.CString(topic)
	defer C.free(unsafe.Pointer(cTopic))
	
	var cCallback C.message_callback
	var userData unsafe.Pointer
	
	if callback != nil {
		// Register the callback
		callbackRegistry.Lock()
		callbackRegistry.callbacks[subscriberID] = callback
		callbackRegistry.Unlock()
		
		// Set the C callback and user data
		cCallback = C.message_callback(C.callbackGateway)
		userData = unsafe.Pointer(cSubscriberID)
	}
	
	success := C.subscribe(cSubscriberID, cTopic, cCallback, userData)
	if !success {
		return errors.New("failed to subscribe")
	}
	
	return nil
}

// Unsubscribe removes a subscription from a topic
// If topic is empty, unsubscribes from all topics
func Unsubscribe(subscriberID string, topic string) error {
	cSubscriberID := C.CString(subscriberID)
	defer C.free(unsafe.Pointer(cSubscriberID))
	
	var cTopic *C.char
	if topic != "" {
		cTopic = C.CString(topic)
		defer C.free(unsafe.Pointer(cTopic))
	}
	
	success := C.unsubscribe(cSubscriberID, cTopic)
	if !success {
		return errors.New("failed to unsubscribe")
	}
	
	// If unsubscribing from all topics, remove the callback
	if topic == "" {
		callbackRegistry.Lock()
		delete(callbackRegistry.callbacks, subscriberID)
		callbackRegistry.Unlock()
	}
	
	return nil
}

// Publish sends a message to a topic
func Publish(topic, message string) error {
	cTopic := C.CString(topic)
	defer C.free(unsafe.Pointer(cTopic))
	
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))
	
	success := C.publish(cTopic, cMessage)
	if !success {
		return fmt.Errorf("failed to publish message to topic '%s'", topic)
	}
	
	return nil
}

// Message represents a pub/sub message
type Message struct {
	Topic   string
	Content string
}

// GetMessage retrieves the next message for a subscriber
// If topic is empty, gets the next message from any topic
func GetMessage(subscriberID string, topic string) (*Message, error) {
	cSubscriberID := C.CString(subscriberID)
	defer C.free(unsafe.Pointer(cSubscriberID))
	
	var cTopic *C.char
	if topic != "" {
		cTopic = C.CString(topic)
		defer C.free(unsafe.Pointer(cTopic))
	}
	
	// Allocate buffers for the output
	cOutTopic := (*C.char)(C.malloc(C.size_t(MaxTopicSize)))
	defer C.free(unsafe.Pointer(cOutTopic))
	
	cOutMessage := (*C.char)(C.malloc(C.size_t(MaxMessageSize)))
	defer C.free(unsafe.Pointer(cOutMessage))
	
	success := C.get_next_message(
		cSubscriberID,
		cTopic,
		cOutTopic,
		C.size_t(MaxTopicSize),
		cOutMessage,
		C.size_t(MaxMessageSize),
	)
	
	if !success {
		return nil, errors.New("no messages available")
	}
	
	return &Message{
		Topic:   C.GoString(cOutTopic),
		Content: C.GoString(cOutMessage),
	}, nil
}

// HasMessages checks if there are any messages available for a subscriber
// If topic is empty, checks for messages from any topic
func HasMessages(subscriberID string, topic string) bool {
	cSubscriberID := C.CString(subscriberID)
	defer C.free(unsafe.Pointer(cSubscriberID))
	
	var cTopic *C.char
	if topic != "" {
		cTopic = C.CString(topic)
		defer C.free(unsafe.Pointer(cTopic))
	}
	
	return bool(C.has_messages(cSubscriberID, cTopic))
}
