FROM golang:latest

WORKDIR /go/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o go_chat .

EXPOSE 8080

CMD ["./go_chat"]
