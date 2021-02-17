FROM golang:1.15-alpine

WORKDIR /geo
RUN mkdir saves

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./out/geo ./main

EXPOSE 7089

CMD ["./out/geo"]