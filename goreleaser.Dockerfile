FROM golang:1.14.4

WORKDIR /go/src/albatross
COPY . .

EXPOSE 8080
CMD ["./albatross"]
