FROM gcr.io/distroless/base 
COPY goblackhole /usr/bin/goblackhole
COPY ./config.yaml /etc/goblackhole/config.yaml
CMD ["/usr/bin/goblackhole"]
