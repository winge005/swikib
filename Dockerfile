FROM golang:1.23
WORKDIR /app

COPY . .

RUN go build -o swiki .

ENTRYPOINT [ "/app/swiki" ]
