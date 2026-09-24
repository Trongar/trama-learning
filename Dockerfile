FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/trama .

FROM alpine:3.22
RUN adduser -D -u 65532 -H app && mkdir -p /data && chown app:app /data
WORKDIR /app
COPY --from=build /out/trama /app/trama
VOLUME ["/data"]
ENV TRAMA_DATA_DIR=/data
EXPOSE 8080
USER 65532:65532
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1
ENTRYPOINT ["/app/trama"]
CMD ["serve", "--http=0.0.0.0:8080"]
