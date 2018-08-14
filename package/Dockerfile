FROM ubuntu:18.04
RUN apt-get update && \
    apt-get install -y iptables ca-certificates
ENV RIO_IN_CLUSTER=true
COPY rio /
COPY local-proxy /
CMD ["/rio"]
