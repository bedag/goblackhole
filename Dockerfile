FROM golang AS gobgp
RUN go get github.com/osrg/gobgp/cmd/gobgp

FROM gcr.io/distroless/base 
COPY goblackhole /usr/bin/goblackhole
COPY ./config.yaml /etc/goblackhole/config.yaml
COPY --from=gobgp /go/bin/gobgp /usr/bin/gobgp
CMD ["/usr/bin/goblackhole"]
