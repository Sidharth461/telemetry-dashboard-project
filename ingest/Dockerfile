
FROM golang:1.26 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ingest .


FROM alpine:3.21
WORKDIR /app
COPY --from=build /app/ingest .
EXPOSE 8080
CMD ["./ingest"]