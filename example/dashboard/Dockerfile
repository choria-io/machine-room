FROM golang:1.26 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v -trimpath -ldflags="-s -w" -o /dashboard ./...

FROM almalinux:latest

COPY --from=builder /dashboard /usr/bin/dashboard

CMD /usr/bin/dashboard
