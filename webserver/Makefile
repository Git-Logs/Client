# Some sensible defaults, should be overrided per-project
BINS ?= webserver
PROJ_NAME ?= github-v2
HOST ?= 100.86.85.125

all:
	CGO_ENABLED=0 go build -v
deploy:
	ssh root@$(HOST) "rm -rf ~/github-v2/webserver-deploy && mkdir ~/github-v2/webserver-deploy"

	# Tar all files in the current directory ignoring the ``webserver`` file
	tar -czvf webserver.tar.gz --exclude webserver .

	# Push to the server
	scp webserver.tar.gz root@$(HOST):~/github-v2/webserver-deploy

	# Remove the tarball
	rm -rf webserver.tar.gz

	# Unpack the tarball and build
	ssh root@$(HOST) "cd ~/github-v2/webserver-deploy && tar -xzvf webserver.tar.gz && make"

	# Kill github-api
	ssh root@$(HOST) "systemctl stop github-api"

	# Move the binary to the correct location
	ssh root@$(HOST) "mv ~/github-v2/webserver-deploy/webserver ~/github-v2/webserver"

	# Remove the deploy folder
	ssh root@$(HOST) "rm -rf ~/github-v2/webserver-deploy"

	# Start github-api
	ssh root@$(HOST) "systemctl start github-api"
remote:
	ssh root@$(HOST)
