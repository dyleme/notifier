FROM golang:latest as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/timetable/main.go

FROM alpine:3.14.2
WORKDIR /app
COPY --from=builder ["/app/main", "/app/main"]
CMD ["/app/main"]