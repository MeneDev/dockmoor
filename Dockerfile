FROM registry:2.6.2
RUN apk add --no-cache openssl

RUN mkdir -p /certs

RUN echo $'[ req ]\n\
default_bits       = 2048\n\
distinguished_name = req_distinguished_name\n\
req_extensions = v3_req\n\
[ req_distinguished_name ]\n\
countryName                 = DE\n\
stateOrProvinceName         = TEST\n\
localityName               = TEST\n\
organizationName           = TEST\n\
commonName                 = test\n\
[v3_req]\n\
subjectAltName = IP:172.17.0.5\n\
[san_env]\n\
subjectAltName=${ENV::SAN}\n\
' > /certs/conf.cnf

RUN cat /certs/conf.cnf

RUN echo "subjectAltName = IP:172.17.0.5" > /certs/extfile.cnf

ENV REGISTRY_HTTP_TLS_CERTIFICATE=/certs/registry.crt
ENV REGISTRY_HTTP_TLS_KEY=/certs/registry.key

RUN echo $'#!/bin/sh\n\
set -e\n\
IP="$(/sbin/ifconfig | grep "inet addr" | grep -v 127.0.0.1 | cut -d: -f2 | awk \'{print $1}\')"\n\
SAN="IP:$IP" openssl req -newkey rsa:4096 -nodes -sha256 -keyout /certs/registry.key \
        -x509 -days 365 -out /certs/registry.crt \
        -extensions san_env \
        -subj "/C=DE/ST=test/L=test/O=IT/CN=myregistrydomain.com" \
        -config /certs/conf.cnf \n\
/entrypoint.sh "$@"\n\
' > /testable-entrypoint.sh

RUN chmod +x /testable-entrypoint.sh
RUN cat /entrypoint.sh

ENTRYPOINT ["/testable-entrypoint.sh"]
CMD ["/etc/docker/registry/config.yml"]