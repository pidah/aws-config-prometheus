FROM golang:1.12-alpine as builder

RUN apk --update add gcc libc-dev git 

ADD . /go/src/github.com/pidah/aws-config-prometheus

WORKDIR /go/src/github.com/pidah/aws-config-prometheus

RUN export GO111MODULE=on && go build .

FROM alpine

RUN apk --update add ca-certificates

COPY --from=builder /go/src/github.com/pidah/aws-config-prometheus /

# Create a group and user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Tell docker that all future commands should run as the appuser user
USER appuser

CMD ["/aws-config-prometheus","--logtostderr","-v=4","2>&1"]
