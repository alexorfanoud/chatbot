FROM golang:1.21-alpine

WORKDIR /app
COPY . .

# Live reloading
RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

RUN go mod download
RUN go build -o ./app ./cmd/chat-server/main.go

ENTRYPOINT CompileDaemon -build="go build -o app ./cmd/chat-server/main.go" -command="./app"
