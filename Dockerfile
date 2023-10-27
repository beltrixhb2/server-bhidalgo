FROM golang:latest AS build

WORKDIR /build

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -o server .

FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk add ca-certificates

WORKDIR /app

COPY --from=build /build/server .

RUN chmod +x server 

RUN pwd && find .

CMD ["./server"]
