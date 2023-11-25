ARG BUILDTIME_BASE=golang:1.21.4
ARG RUNTIME_BASE=gcr.io/distroless/static:latest
FROM ${BUILDTIME_BASE} as builder

WORKDIR /go/src/app
ENV CGO_ENABLED=0
ADD . /go/src/app
EXPOSE 3001

RUN go build -ldflags '-s -w' -o /go/bin/ntopng-exporter

FROM ${RUNTIME_BASE}

COPY --from=builder /go/bin/ntopng-exporter /
CMD ["/ntopng-exporter"]
