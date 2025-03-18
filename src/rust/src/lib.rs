use libc::{c_char, c_void};
use once_cell::sync::Lazy;
use std::collections::{HashMap, HashSet, VecDeque};
use std::ffi::{CStr, CString};
use std::sync::Mutex;

// Type for callback function that will be called when a message is published
type MessageCallback = extern "C" fn(*const c_char, *const c_char, *mut c_void);

// Global state for our pub/sub system
static PUBSUB: Lazy<Mutex<PubSubState>> = Lazy::new(|| Mutex::new(PubSubState::new()));

struct CallbackData(*mut c_void);

// Implement Send and Sync for CallbackData
unsafe impl Send for CallbackData {}
unsafe impl Sync for CallbackData {}

struct PubSubState {
    // Map of topic to set of subscriber IDs
    topics: HashMap<String, HashSet<String>>,
    // Map of subscriber ID to callback function and user data
    callbacks: HashMap<String, (MessageCallback, CallbackData)>,
    // Queue of messages for subscribers without callbacks
    message_queues: HashMap<String, VecDeque<(String, String)>>,
}

impl PubSubState {
    fn new() -> Self {
        PubSubState {
            topics: HashMap::new(),
            callbacks: HashMap::new(),
            message_queues: HashMap::new(),
        }
    }
}

// Helper function to convert C string to Rust string
fn c_str_to_string(c_str: *const c_char) -> String {
    let c_str = unsafe { CStr::from_ptr(c_str) };
    c_str.to_string_lossy().into_owned()
}

#[no_mangle]
pub extern "C" fn subscribe(
    subscriber_id: *const c_char,
    topic: *const c_char,
    callback: Option<MessageCallback>,
    user_data: *mut c_void,
) -> bool {
    if subscriber_id.is_null() || topic.is_null() {
        return false;
    }

    let subscriber_id = c_str_to_string(subscriber_id);
    let topic = c_str_to_string(topic);

    let mut state = PUBSUB.lock().unwrap();

    // Create topic if it doesn't exist
    let subscribers = state
        .topics
        .entry(topic.clone())
        .or_insert_with(HashSet::new);
    subscribers.insert(subscriber_id.clone());

    // Store callback if provided
    if let Some(cb) = callback {
        state
            .callbacks
            .insert(subscriber_id.clone(), (cb, CallbackData(user_data)));
    } else {
        // Initialize message queue for this subscriber if no callback
        state
            .message_queues
            .entry(subscriber_id.clone())
            .or_insert_with(VecDeque::new);
    }

    true
}

#[no_mangle]
pub extern "C" fn unsubscribe(subscriber_id: *const c_char, topic: *const c_char) -> bool {
    if subscriber_id.is_null() {
        return false;
    }

    let subscriber_id = c_str_to_string(subscriber_id);
    let mut state = PUBSUB.lock().unwrap();

    if topic.is_null() {
        // Unsubscribe from all topics
        for (_, subscribers) in state.topics.iter_mut() {
            subscribers.remove(&subscriber_id);
        }

        // Remove callback and message queue
        state.callbacks.remove(&subscriber_id);
        state.message_queues.remove(&subscriber_id);
    } else {
        // Unsubscribe from specific topic
        let topic = c_str_to_string(topic);
        if let Some(subscribers) = state.topics.get_mut(&topic) {
            subscribers.remove(&subscriber_id);
        }
    }

    true
}

#[no_mangle]
pub extern "C" fn publish(topic: *const c_char, message: *const c_char) -> bool {
    if topic.is_null() || message.is_null() {
        return false;
    }

    let topic_str = c_str_to_string(topic);
    let message_str = c_str_to_string(message);

    let mut state = PUBSUB.lock().unwrap();

    // Check if topic exists
    let subscribers = match state.topics.get(&topic_str) {
        Some(subs) => subs.clone(), // Clone the subscribers to avoid borrow issues
        None => return false,       // Topic doesn't exist
    };

    // Convert topic and message to C strings once
    let topic_c_str = CString::new(topic_str.clone()).unwrap();
    let message_c_str = CString::new(message_str.clone()).unwrap();

    // Process each subscriber
    for subscriber_id in subscribers {
        // If subscriber has a callback, invoke it
        if let Some((callback, user_data)) = state.callbacks.get(&subscriber_id) {
            let cb = *callback;
            cb(topic_c_str.as_ptr(), message_c_str.as_ptr(), user_data.0);
        } else {
            // Otherwise, queue the message
            if let Some(queue) = state.message_queues.get_mut(&subscriber_id) {
                queue.push_back((topic_str.clone(), message_str.clone()));
            }
        }
    }

    true
}

#[no_mangle]
pub extern "C" fn get_next_message(
    subscriber_id: *const c_char,
    topic: *const c_char,
    out_topic: *mut c_char,
    out_topic_size: usize,
    out_message: *mut c_char,
    out_message_size: usize,
) -> bool {
    if subscriber_id.is_null() {
        return false;
    }

    let subscriber_id = c_str_to_string(subscriber_id);
    let mut state = PUBSUB.lock().unwrap();

    // Get the message queue for this subscriber
    if let Some(queue) = state.message_queues.get_mut(&subscriber_id) {
        // If topic is specified, find a message for that topic
        if !topic.is_null() {
            let topic_str = c_str_to_string(topic);

            // Find the index of the first message for this topic
            if let Some(index) = queue.iter().position(|(t, _)| t == &topic_str) {
                let (topic, message) = queue.remove(index).unwrap();

                // Copy topic and message to output buffers if provided
                if !out_topic.is_null() && out_topic_size > 0 {
                    let bytes_to_copy = std::cmp::min(topic.len(), out_topic_size - 1);
                    unsafe {
                        std::ptr::copy_nonoverlapping(
                            topic.as_ptr(),
                            out_topic as *mut u8,
                            bytes_to_copy,
                        );
                        *out_topic.add(bytes_to_copy) = 0; // Null terminator
                    }
                }

                if !out_message.is_null() && out_message_size > 0 {
                    let bytes_to_copy = std::cmp::min(message.len(), out_message_size - 1);
                    unsafe {
                        std::ptr::copy_nonoverlapping(
                            message.as_ptr(),
                            out_message as *mut u8,
                            bytes_to_copy,
                        );
                        *out_message.add(bytes_to_copy) = 0; // Null terminator
                    }
                }

                return true;
            }
        } else if let Some((topic, message)) = queue.pop_front() {
            // Get the next message regardless of topic
            // Copy topic and message to output buffers if provided
            if !out_topic.is_null() && out_topic_size > 0 {
                let bytes_to_copy = std::cmp::min(topic.len(), out_topic_size - 1);
                unsafe {
                    std::ptr::copy_nonoverlapping(
                        topic.as_ptr(),
                        out_topic as *mut u8,
                        bytes_to_copy,
                    );
                    *out_topic.add(bytes_to_copy) = 0; // Null terminator
                }
            }

            if !out_message.is_null() && out_message_size > 0 {
                let bytes_to_copy = std::cmp::min(message.len(), out_message_size - 1);
                unsafe {
                    std::ptr::copy_nonoverlapping(
                        message.as_ptr(),
                        out_message as *mut u8,
                        bytes_to_copy,
                    );
                    *out_message.add(bytes_to_copy) = 0; // Null terminator
                }
            }

            return true;
        }
    }

    false
}

#[no_mangle]
pub extern "C" fn has_messages(subscriber_id: *const c_char, topic: *const c_char) -> bool {
    if subscriber_id.is_null() {
        return false;
    }

    let subscriber_id = c_str_to_string(subscriber_id);
    let state = PUBSUB.lock().unwrap();

    if let Some(queue) = state.message_queues.get(&subscriber_id) {
        if topic.is_null() {
            // Check if there are any messages
            return !queue.is_empty();
        } else {
            // Check if there are messages for the specific topic
            let topic_str = c_str_to_string(topic);
            return queue.iter().any(|(t, _)| t == &topic_str);
        }
    }

    false
}
