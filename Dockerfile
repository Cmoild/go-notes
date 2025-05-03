FROM golang:1.24-alpine

# Установим psql
RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN chmod +x run.sh

RUN go build -o main ./src/main.go

EXPOSE 8080

CMD ["./run.sh"]
