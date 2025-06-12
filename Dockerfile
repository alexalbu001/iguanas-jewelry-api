FROM golang:1.24-alpine

WORKDIR /app

COPY . .

RUN go mod download && go mod verify

RUN go build -o main cmd/api/main.go

EXPOSE 8080

CMD ["./main"]