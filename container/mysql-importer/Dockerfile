FROM golang:alpine AS build-env

# Install some dependencies needed to build the project
RUN apk add ca-certificates git && rm -rf /var/cache/apk/*
WORKDIR /app
ADD . /app
ARG git_revision
RUN CGO_ENABLED=0 go build -ldflags="-w -s -X main.Revision=$git_revision" -o goapp

FROM alpine
WORKDIR /app
COPY --from=build-env /app/goapp /tmp/goapp
COPY ./run.sh /tmp/run
COPY ./loop.sh /tmp/loop.sh
RUN /bin/chmod a+x /tmp/goapp && chmod a+x /tmp/run && chmod a+x /tmp/loop.sh
ARG git_revision
ENV GIT_REVISION $git_revision
CMD ["/tmp/loop.sh"]
