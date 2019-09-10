.PHONY: build run deploy vendor
build:
	docker build . -t quay.io/fofgaming/discordstats
run:
	docker run --rm -e DISCORD_BOT_TOKEN -e DB_CONNECTION_STRING -e ENVIRONMENT -e ROLLBAR_TOKEN quay.io/fofgaming/discordstats
deploy:
	docker push quay.io/fofgaming/discordstats
vendor:
	GO111MODULE=on go mod vendor
