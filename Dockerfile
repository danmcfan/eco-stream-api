FROM golang:1.22

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN go build -v -o ./bin/server ./cmd/server/main.go

CMD ["./bin/server"]
