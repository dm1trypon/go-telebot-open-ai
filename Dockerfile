FROM golang:1.20

WORKDIR /usr/src/tbotopenai

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/tbotopenai ./cmd/gotbotopenai/main.go

EXPOSE 80

CMD ["tbotopenai"]
