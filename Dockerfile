FROM scratch
COPY goblackhole /usr/bin/goblackhole
ENTRYPOINT ["/usr/bin/goblackhole"]