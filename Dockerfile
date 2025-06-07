ARG certs_image=alpine:3.21.3
ARG root_image=alpine:3.21.3
ARG TARGETARCH
FROM ${certs_image} AS certs
RUN apk --no-cache add ca-certificates

FROM ${root_image}
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY bin/defectdojo-exporter-linux-${TARGETARCH} ./defectdojo-exporter
EXPOSE 8429
ENTRYPOINT ["./defectdojo-exporter"]
