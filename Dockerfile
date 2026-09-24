FROM golang:1.26.6-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/trama .

FROM alpine:3.22
RUN adduser -D -u 65532 -H app
WORKDIR /app
COPY --from=build /out/trama /app/trama
ENV EDG_ADDR=:8080
EXPOSE 8080
USER 65532:65532
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1
ENTRYPOINT ["/app/trama"]
