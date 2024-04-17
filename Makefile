CDN_PATH := /silverpelt/cdn/ibl

restartbot:
	cd bot && cargo sqlx prepare
	cd bot && cargo build --release
	make restartbot_nobuild

restartbot_nobuild:
	sudo systemctl stop github-bot
	sleep 3 # Give time for it to stop
	mkdir -p dist
	cp -v bot/target/release/bot dist/bot
	sudo systemctl start github-bot

restartwebserver:
	cd webserver && CGO_ENABLED=0 go build -v
	make restartwebserver_nobuild

restartwebserver_nobuild:
	sudo systemctl stop github-api
	sleep 3 # Give time for it to stop
	mkdir -p dist
	cp -v webserver/webserver dist/webserver
	sudo systemctl start github-api