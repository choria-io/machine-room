FROM golang:1.26 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v -trimpath -ldflags="-s -w" -o /saas-manager ./...

FROM almalinux:latest

RUN mkdir /etc/saas-manager

COPY --from=builder /saas-manager /usr/bin/saas-manager

CMD /usr/bin/saas-manager run --config /etc/saas-manager/agent.cfg
