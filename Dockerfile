FROM gcr.io/distroless/static-debian12:nonroot

COPY relayer_exporter /usr/bin/relayer_exporter

# metrics server
EXPOSE 8008

ENTRYPOINT [ "/usr/bin/relayer_exporter" ]

CMD [ "--help" ]
