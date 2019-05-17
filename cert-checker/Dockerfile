FROM alpine:3.9
RUN apk add -U --no-cache openssl
COPY cert-checker/init.sh /
CMD ["/init.sh"]
