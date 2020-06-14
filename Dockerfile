ARG GOLANG_VERSION=1.14.2

FROM golang:${GOLANG_VERSION}-stretch as base
LABEL maintainer="Nick Calibey"
WORKDIR /github.com/ncalibey/ls-todo

COPY ./go.mod go.mod
COPY ./go.sum go.sum
RUN go mod download
# For local development, it's usually faster to just use vendoring.
#ENV GOFLAGS=-mod=vendor
#COPY ./vendor vendor

##############################################################################
# Builder Stage ##############################################################
FROM base as builder
COPY ./cmd cmd
COPY ./internal internal
RUN go build -o /bin/todo-server cmd/main/main.go

##############################################################################
# Release Stage ##############################################################

# By using a release stage that is based off this image, our final binary is
# much smaller.
FROM debian:stretch-slim as release
EXPOSE 8080

COPY --from=builder /bin/todo-server /bin/todo-server
CMD ["/bin/todo-server"]
