FROM golang:1.11 as build

RUN mkdir -p /go/src/github.com/stationa/xgeo
ADD . /go/src/github.com/stationa/xgeo/
WORKDIR /go/src/github.com/stationa/xgeo
RUN make release

FROM scratch
COPY --from=build /go/bin/xgeo /usr/local/bin/xgeo
ENTRYPOINT ["/usr/local/bin/xgeo"]
