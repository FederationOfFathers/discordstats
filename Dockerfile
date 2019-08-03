FROM golang:1.12-alpine
ENV GO111MODULE=on
ENV CGO_ENABLED=0
RUN apk update && apk add git
WORKDIR /app
COPY . /app
RUN cd /app && go build
ENTRYPOINT ["/app/discordstats"]