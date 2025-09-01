# Frontend builder stage
FROM --platform=linux/amd64 mirror.gcr.io/library/node:20 AS frontend-builder

WORKDIR /app/web

# Copy package.json
COPY web/package.json ./

# Force clean installation
RUN npm install --force --no-package-lock

# Copy frontend source
COPY web/ ./

# Build frontend
RUN npm run build

# Go builder stage  
FROM --platform=linux/amd64 mirror.gcr.io/library/golang:1.24.6-alpine AS go-builder

WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy Go source code
COPY . .

# Build Go backend
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Final minimal image
FROM --platform=linux/amd64 mirror.gcr.io/library/alpine:3.20

WORKDIR /app

# Copy binary from Go builder
COPY --from=go-builder /app/app .

# Copy frontend build output from frontend builder
COPY --from=frontend-builder /app/web/dist ./web/dist

EXPOSE 8080
CMD ["./app"]