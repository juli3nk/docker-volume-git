FROM golang:1.12-alpine3.9 AS builder

RUN apk --update add \
		ca-certificates \
		gcc \
		git \
		musl-dev

COPY go.mod go.sum /go/src/github.com/kassisol/docker-volume-git/
WORKDIR /go/src/github.com/kassisol/docker-volume-git

ENV GO111MODULE on
RUN go mod download

COPY . .

RUN go build -ldflags "-linkmode external -extldflags -static -s -w" -o /tmp/docker-volume-git \
	&& strip --strip-all /tmp/docker-volume-git


FROM alpine

RUN apk --update --no-cache add \
		ca-certificates \
	&& mkdir -p /var/lib/docker/volumes /var/lib/docker/state

COPY --from=builder /tmp/docker-volume-git /docker-volume-git

ENTRYPOINT ["/docker-volume-git"]
