.PHONY: all clean rust go

# Default target
all: rust go

# Create target directories
target/release:
	mkdir -p target/release

# Build Rust library
rust: target/release
	@echo "Building Rust library..."
	cd src/rust && cargo build --release
	cp src/rust/target/release/libpubsub_core.* target/release/

# Build Go application
go: rust
	@echo "Building Go application..."
	cd src/go && \
	go mod init github.com/jbrinkman/go-rust-ffi/go && \
	go mod tidy && \
	CGO_LDFLAGS="-L../../target/release" go build -o ../../target/release/pubsub_example

# Clean build artifacts
clean:
	rm -rf target
	cd src/rust && cargo clean
	cd src/go && rm -f go.mod go.sum

# Help message
help:
	@echo "Available targets:"
	@echo "  all    - Build both Rust library and Go application (default)"
	@echo "  rust   - Build only the Rust library"
	@echo "  go     - Build the Go application (also builds Rust if needed)"
	@echo "  clean  - Remove all build artifacts"
	@echo "  help   - Show this help message"

# Print run instructions
.PHONY: run-instructions
run-instructions:
	@echo "Build complete. Run with: ./target/release/pubsub_example"

# Add run instructions after successful build
all go: run-instructions
