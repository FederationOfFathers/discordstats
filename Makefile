build:
	docker build . -t fofgaming/discordstats
run:
	docker run --rm -e DISCORD_BOT_TOKEN -e DB_CONNECTION_STRING -e ENVIRONMENT -e ROLLBAR_TOKEN fofgaming/discordstats
deploy:
	docker push fofgaming/discordstats
vendor:
	GO111MODULE=on go mod vendor