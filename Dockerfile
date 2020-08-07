FROM golang:1.14.4

WORKDIR /go/src/albatross
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080
CMD ["albatross"]
