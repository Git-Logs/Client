RUSTFLAGS_LOCAL="-C target-cpu=native $(RUSTFLAGS) -C link-arg=-fuse-ld=lld"
CARGO_TARGET_GNU_LINKER="x86_64-unknown-linux-gnu-gcc"
DATABASE_URL=postgresql:///github

# Some sensible defaults, should be overrided per-project
BINS ?= bot
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
	# Kill github-bot
	ssh root@$(HOST) "systemctl stop github-bot"

	@for bin in $(BINS) ; do \
		echo "Pushing $$bin to $(HOST):${PROJ_NAME}/$$bin"; \
		scp -C target/x86_64-unknown-linux-gnu/release/$$bin root@$(HOST):${PROJ_NAME}/$$bin; \
	done

	# Start github-bot
	ssh root@$(HOST) "systemctl start github-bot"

	DATABASE_URL=$(DATABASE_URL) cargo test ${ARGS}

	ssh root@$(HOST) "mkdir -p /iblseeds/apiBindings"

	scp -r .generated root@${HOST}:/iblseeds/apiBindings/

	ssh root@$(HOST) "rm /iblseeds/apiBindings/*.ts && cp /iblseeds/apiBindings/.generated/*.ts /iblseeds/apiBindings/ && rm -rf /iblseeds/apiBindings/.generated"

	@# Remove the .generated folder
	rm -rf .generated

run:
	-mv -vf bot.new bot # If it exists
	./bot
remote:
	ssh root@$(HOST)
