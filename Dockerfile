FROM golang:1.14-alpine AS driver_build

RUN apk add git

WORKDIR /go/src/github.com/simonasr/docker-jsonudp-log-driver

RUN go get github.com/docker/docker/api/types/plugins/logdriver
RUN go get github.com/docker/docker/daemon/logger
RUN go get github.com/docker/docker/daemon/logger/loggerutils
RUN go get github.com/docker/go-plugins-helpers/sdk
RUN go get github.com/pkg/errors
RUN go get github.com/Sirupsen/logrus
RUN go get github.com/tonistiigi/fifo
RUN go get github.com/gogo/protobuf/io
RUN go get github.com/snowzach/rotatefilehook

COPY . /go/src/github.com/simonasr/docker-jsonudp-log-driver
RUN go get
RUN go build --ldflags '-extldflags "-static"' -o /usr/bin/docker-jsonudp-log-driver

FROM alpine:3.7
RUN apk --no-cache add ca-certificates
COPY --from=driver_build /usr/bin/docker-jsonudp-log-driver /usr/bin/
WORKDIR /usr/bin/
ENTRYPOINT ["/usr/bin/docker-jsonudp-log-driver"]
