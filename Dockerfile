# Dependencies

## Make sure your machine has >= 1 GB of RAM.

## Make sure go version >= 1.3.3 is installed and set up
FROM golang:1.4
MAINTAINER Eris Industries <contact@erisindustries.com>

### The base image kills /var/lib/apt/lists/*.
RUN apt-get update
RUN apt-get install -y \
  libgmp3-dev

## Copy In the Good Stuff
RUN go get github.com/eris-ltd/epm-go/cmd/epm
COPY start.sh /

## How Does It Run?
EXPOSE 15254 15255
CMD ["/start.sh"]