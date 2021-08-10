FROM golang AS gobgp
RUN go get github.com/osrg/gobgp/cmd/gobgp

FROM golang AS gobgp-exporter
RUN cd $GOPATH/src \
    && mkdir -p github.com/greenpau \
    && cd github.com/greenpau \ 
    && git clone https://github.com/greenpau/gobgp_exporter.git \
    && cd gobgp_exporter \
    && CGO_ENABLED=0 make

FROM gcr.io/distroless/base 
COPY goblackhole /usr/bin/goblackhole
COPY ./config.yaml /etc/goblackhole/config.yaml
COPY --from=gobgp-exporter /go/src/github.com/greenpau/gobgp_exporter/bin/gobgp-exporter /usr/bin/gobgp-exporter
COPY --from=gobgp /go/bin/gobgp /usr/bin/gobgp
CMD ["/usr/bin/goblackhole"]
