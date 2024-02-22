FROM golang:1.22

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./main.go ./main.go
COPY ./internal ./internal

RUN go build -v -o ./bin/main ./main.go

CMD ["./bin/main"]
