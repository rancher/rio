FROM alpine:3.9
RUN apk -U --no-cache add openssl bash
COPY cert-check.sh /
RUN chmod +x /cert-check.sh
CMD ["/cert-check.sh", "/etc/registry/tls.crt"]