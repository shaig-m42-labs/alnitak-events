FROM golang:1.26-alpine AS build
WORKDIR /workspace
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /out/events ./cmd/events

FROM alpine:3.21
RUN adduser -D -g '' appuser
USER appuser
WORKDIR /app
COPY --from=build /out/events /app/events
EXPOSE 8083
ENTRYPOINT ["/app/events"]
