FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache arp-scan
COPY --from=build /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
