# Go-Rust FFI Demo

This project demonstrates Foreign Function Interface (FFI) between Go and Rust using a publish-subscribe pattern. It showcases how to create a Rust library that can be called from Go, with proper memory management and thread safety.

## Project Structure

```
.
├── src/
│   ├── go/       # Go code that calls into the Rust library
│   └── rust/     # Rust library with FFI exports
├── build.sh      # Build script
└── Makefile      # Makefile for building the project
```

## Features

- Thread-safe publish-subscribe system implemented in Rust
- FFI bindings to call Rust from Go
- Support for callbacks from Rust to Go
- Message queuing for subscribers without callbacks
- Proper memory management across language boundaries

## Requirements

- Rust (stable channel)
- Go 1.22+ (currently supported versions as of March 2025)
- C compiler (for cgo)

## Building

To build the project, run:

```bash
make
```

Or use the build script:

```bash
./build.sh
```

## Usage

The library provides the following core functions:

- `subscribe`: Subscribe to a topic with an optional callback
- `unsubscribe`: Unsubscribe from a topic
- `publish`: Publish a message to a topic
- `get_next_message`: Get the next message for a subscriber
- `has_messages`: Check if a subscriber has pending messages

See the Go examples in `src/go` for usage patterns.

## Thread Safety

The Rust library uses `Mutex` and thread-safe wrappers to ensure that the pub-sub system can be safely used from multiple threads, both in Rust and when called from Go.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
