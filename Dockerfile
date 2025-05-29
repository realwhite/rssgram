FROM --platform=linux/arm64 golang:1.23-alpine as builder
WORKDIR /build
RUN apk add --no-cache gcc musl-dev
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o /rssgramm ./cmd/main.go

FROM --platform=linux/arm64 alpine:3
RUN apk add --no-cache tzdata
COPY --from=builder /rssgramm /bin/rssgramm
ENTRYPOINT ["/bin/rssgramm"]


