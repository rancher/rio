FROM golang:1.10 AS cni
RUN mkdir -p $GOPATH/src/github.com/containernetworking && \
    cd $GOPATH/src/github.com/containernetworking && \
    git clone https://github.com/ibuildthecloud/plugins.git && \
    cd plugins && \
    git checkout 299b021fdcb5533d5babc40006d2965a8563066a
RUN go build -ldflags -s -o /usr/bin/cni github.com/containernetworking/plugins

FROM rio-build AS ubuntu

RUN mkdir /usr/src/image && \
    tar xvf /usr/src/buildroot/output/images/rootfs.tar -C /usr/src/image ./etc && \
    tar xvf /usr/src/buildroot/output/images/rootfs.tar -C /usr/src/image ./usr --strip-components=2

RUN curl -fL -o /usr/src/image/bin/runc https://github.com/opencontainers/runc/releases/download/v1.0.0-rc5/runc.amd64
RUN chmod +x /usr/src/image/bin/runc && strip --strip-debug /usr/src/image/bin/runc

RUN curl -fL -o /usr/src/containerd.tgz https://github.com/containerd/containerd/releases/download/v1.1.1/containerd-1.1.1.linux-amd64.tar.gz
RUN tar xvf /usr/src/containerd.tgz -C /usr/src/image/bin bin/containerd-shim bin/containerd --strip-components=1

RUN mkdir -p /usr/src/image/share/cni/bin
COPY --from=cni /usr/bin/cni /usr/src/image/share/cni/bin
RUN cd /usr/src/image/share/cni/bin && \
    for i in ./bridge ./flannel ./host-local ./loopback ./portmap; do \
        ln -s cni $i \
    ;done

RUN mkdir -p /usr/src/image/etc/containerd
COPY config.toml.tmpl /usr/src/image/etc/containerd/config.toml.tmpl
COPY init /usr/src/image/init

RUN mkdir -p /usr/src/image/etc/cni/net.d
COPY 10-flannel.conflist /usr/src/image/etc/cni/net.d

RUN rm -rf /usr/src/image/etc/init.d
COPY inittab /usr/src/image/etc/inittab

RUN mkdir -p /usr/src/image/etc/kube-flannel
COPY net-conf.json /usr/src/image/etc/kube-flannel/

COPY agent /usr/src/image/bin/agent

RUN cd /usr/src/image && \
    mkdir -p lib/modules lib/firmware && \
    rm -rf dev etc/network share/udhcpc

FROM scratch
COPY --from=ubuntu /usr/src/image/ /
CMD ["/bin/bash"]
