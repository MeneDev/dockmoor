FROM registry

COPY registry_data /var/lib/registry

RUN mkdir /certs

COPY registry.localhost.key /certs/registry.key
COPY registry.localhost.crt /certs/registry.crt

ENV REGISTRY_HTTP_TLS_CERTIFICATE=/certs/registry.crt
ENV REGISTRY_HTTP_TLS_KEY=/certs/registry.key