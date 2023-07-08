FROM golang:1.20-alpine

WORKDIR /bob-app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/local/bin/bob

ENV DEV_MODE=true
ENV MONGODB_DB_NAME=bob
ENV MONGODB_CONNECTION_URL=""
ENV HOST=0.0.0.0
ENV PORT=8080

EXPOSE ${PORT}

CMD ["/usr/local/bin/bob"]
