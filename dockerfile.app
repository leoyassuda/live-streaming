FROM golang:1.23-alpine

RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY go.mod .
COPY main.go .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]