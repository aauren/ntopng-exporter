FROM golang:1.17.6 as build

WORKDIR /go/src/app
ENV CGO_ENABLED=0
ADD . /go/src/app
EXPOSE 3001

RUN go build -ldflags '-s -w' -o /go/bin/ntopng-exporter

FROM gcr.io/distroless/static

COPY --from=build /go/bin/ntopng-exporter /
CMD ["/ntopng-exporter"]
