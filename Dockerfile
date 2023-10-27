FROM golang:1.21.1

WORKDIR /app

COPY . .

RUN go build -o server .

EXPOSE 8080

CMD ["./main_opensky"]
