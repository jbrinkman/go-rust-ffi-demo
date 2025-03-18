#!/bin/bash
set -e

# Create target directories if they don't exist
mkdir -p target/release

echo "Building Rust library..."
cd src/rust
cargo build --release
cd ../..

# Copy the compiled library to the target directory
cp src/rust/target/release/libpubsub_core.* target/release/

echo "Building Go application..."
cd src/go
go mod init github.com/jbrinkman/go-rust-ffi/go
go mod tidy
CGO_LDFLAGS="-L../../target/release" go build -o ../../target/release/pubsub_example

echo "Build complete. Run with: ./target/release/pubsub_example"
