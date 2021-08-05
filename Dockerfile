FROM scratch
COPY goblackhole /usr/bin/goblackhole
COPY ./config.yaml /etc/goblackhole/config.yaml
ENTRYPOINT ["/usr/bin/goblackhole"]