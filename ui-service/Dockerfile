# Build the UI service binary and frontend
FROM docker.io/node:20 AS frontend-builder
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM docker.io/golang:1.23 AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy frontend build into static directory
COPY --from=frontend-builder /app/build ./static
RUN CGO_ENABLED=0 GOOS=linux go build -a -o ui-service ./cmd/main.go

# Use distroless as minimal base image
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/ui-service .
COPY --from=builder /workspace/static ./static
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/ui-service"]