FROM scratch

COPY relayer_exporter /usr/bin/relayer_exporter

USER 1000:1000

# metrics server
EXPOSE 8008

ENTRYPOINT [ "/usr/bin/relayer_exporter" ]

CMD [ "--help" ]
