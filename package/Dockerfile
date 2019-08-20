FROM alpine:3.9
RUN apk -U --no-cache add ca-certificates
COPY bin/rio-controller bin/rio /usr/bin/
CMD ["rio-controller"]
