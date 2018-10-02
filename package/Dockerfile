FROM ubuntu:18.04
RUN apk -U --no-cache add ca-certificates
ENV RIO_IN_CLUSTER=true
COPY rio /
CMD ["/rio"]
