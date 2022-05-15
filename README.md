# GitUp

[![GitHub Action][0]][1]

The Go-based static blog generator. This is the simple static website generator
based on the Go and Markdown.

```bash
> gitup -vv clone YOUR_REMOTE_REPOSITORY
```

## Dockerfile

The following is the sample Dockerfile to build the static HTML webpage from the current
folder which contains the `.gitup.yml` setting file.

```dockerfile
FROM golang:1.18-alpine

ARG VERSION="v0.2.3"
RUN go install github.com/cmj0121/gitup@$VERSION

WORKDIR /src
COPY . .

RUN gitup -vv clone file:// && \
	# change the permission of the static file
	chmod 644 build/*

# ================ #
# the final stage  #
# ================ #
FROM nginx:1.21.6-alpine-perl

# expose the port 80
EXPOSE 80/tcp

RUN rm -rf /usr/share/nginx/html/*
COPY --from=builder /src/build/ /usr/share/nginx/html/
```

[0]: https://github.com/cmj0121/gitup/actions/workflows/test.yml/badge.svg
[1]: https://github.com/cmj0121/gitup/actions/workflows/test.yml
