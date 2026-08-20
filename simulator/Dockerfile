
FROM golang:1.26 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o simulator .

FROM alpine:3.21
WORKDIR /app
COPY --from=build /app/simulator .
CMD ["./simulator"]