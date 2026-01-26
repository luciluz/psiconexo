FROM golang:alpine

RUN apk add --no-cache build-base
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o psiconexo-api main.go

EXPOSE 8080
CMD ["./psiconexo-api"]