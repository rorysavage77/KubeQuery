# Remove frontend build steps and just copy static HTML
FROM golang:1.24 as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ui-service ./cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/ui-service /ui-service
COPY --from=builder /app/static /static
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/ui-service"]