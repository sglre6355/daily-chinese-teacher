FROM golang AS builder
WORKDIR /app
# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download
# Copy remaining source files
COPY . .
# Build the application
RUN go build -o daily-chinese-teacher .

FROM gcr.io/distroless/cc
COPY --from=builder /app/daily-chinese-teacher /
CMD ["/daily-chinese-teacher"]
