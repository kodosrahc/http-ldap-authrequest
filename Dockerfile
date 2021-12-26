FROM golang:1.16-alpine3.14 as builder
RUN apk add --no-cache git build-base make pkgconf openssl-dev openssl-libs-static zlib-static libxml2-dev vim
ENV GO111MODULE=on
ENV GOCACHE=/tmp/.go-cache
COPY ./ /app/
WORKDIR /app
RUN make

FROM alpine:3.14
COPY --from=builder /app/bin/http-ldap-authrequest /app/
ENTRYPOINT ["/app/http-ldap-authrequest"]
USER 1000:1000
WORKDIR /app
