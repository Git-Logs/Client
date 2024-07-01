CDN_PATH := /silverpelt/cdn/ibl

restartbot:
	cd bot && cargo sqlx prepare
	cd bot && cargo build --release
	make restartbot_nobuild

restartbot_nobuild:
	systemctl stop github-bot
	sleep 3 # Give time for it to stop
	mkdir -p dist
	cp -v bot/target/release/bot dist/bot
	systemctl start github-bot

restartwebserver:
	cd webserver && CGO_ENABLED=0 go build -v
	make restartwebserver_nobuild

restartwebserver_nobuild:
	systemctl stop github-api
	sleep 3 # Give time for it to stop
	mkdir -p dist
	cp -v webserver/webserver dist/webserver
	systemctl start github-api
