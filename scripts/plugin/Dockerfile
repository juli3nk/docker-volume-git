FROM alpine

RUN apk --update --no-cache add \
		ca-certificates \
	&& mkdir -p /var/lib/docker/volumes /var/lib/docker/state

COPY build/docker-volume-git /docker-volume-git

ENTRYPOINT ["/docker-volume-git"]
