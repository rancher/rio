FROM ubuntu:18.04
RUN apt-get update && \
    apt-get install -y iptables
ENV RIO_IN_CLUSTER=true
COPY rio /
COPY local-proxy /
CMD ["/rio"]
