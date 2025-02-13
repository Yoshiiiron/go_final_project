FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o go_final_project .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app .

EXPOSE 7540

CMD ["./go_final_project"]