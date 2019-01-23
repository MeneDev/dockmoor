FROM registry

COPY registry_data /var/lib/registry

RUN mkdir /certs
RUN mkdir /auth

COPY certs/pki/private/registry.localhost.key /certs/registry.key
COPY certs/pki/issued/registry.localhost.crt /certs/registry.crt
COPY auth/.htpasswd /auth/.htpasswd

ENV REGISTRY_HTTP_TLS_CERTIFICATE=/certs/registry.crt
ENV REGISTRY_HTTP_TLS_KEY=/certs/registry.key