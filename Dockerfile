ARG GO_VERSION=1.25.4

FROM golang:${GO_VERSION} AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/app

FROM gcr.io/distroless/static-debian13:nonroot AS runtime

COPY --from=builder /bin/app /app

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD ["/app","-healthcheck"]

USER nonroot:nonroot

ENTRYPOINT ["/app"]
