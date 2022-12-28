include .env

RUSTFLAGS_LOCAL="-C target-cpu=native $(RUSTFLAGS) -C link-arg=-fuse-ld=lld"
CARGO_TARGET_GNU_LINKER="x86_64-unknown-linux-gnu-gcc"

# Some sensible defaults, should be overrided per-project
BIN_NAME ?= bot
PROJ_NAME ?= github-v2
HOST ?= 100.86.85.125

all: 
	@make cross
cross:
	DATABASE_URL=$(DATABASE_URL) CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER=$(CARGO_TARGET_GNU_LINKER) cargo build --target=x86_64-unknown-linux-gnu --release ${ARGS}
selfhostbuild:
	DATABASE_URL=$(DATABASE_URL) cargo build --release ${ARGS}
selfhostrun:
	DATABASE_URL=$(DATABASE_URL) cargo run --release ${ARGS}
push:
	@for bin in $(BIN_NAME) ; do \
		echo "Pushing $$bin to $(HOST):${PROJ_NAME}/$$bin/$$bin.new"; \
		scp -C target/x86_64-unknown-linux-gnu/release/$$bin root@$(HOST):${PROJ_NAME}/$$bin/$$bin.new; \
	done
run:
	-mv -vf $(BIN_NAME).new $(BIN_NAME) # If it exists
	./$(BIN_NAME)
remote:
	ssh root@$(HOST)
