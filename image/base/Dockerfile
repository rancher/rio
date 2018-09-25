FROM ubuntu:18.04 AS ubuntu
RUN yes | unminimize
RUN apt-get update
RUN apt-get install -y \
    build-essential \
    gcc \
    g++ \
    rsync \
    bc \
    wget \
    curl \
    ca-certificates \
    ncurses-dev \
    python \
    unzip
RUN mkdir /usr/src/buildroot
RUN curl -fL https://buildroot.org/downloads/buildroot-2018.02.5.tar.bz2 | tar xvjf - -C /usr/src/buildroot --strip-components=1
RUN curl -fL https://storage.googleapis.com/buildroot-cache/2018.02.2.tar.gz | tar xvzf - -C /usr/src/buildroot
WORKDIR /usr/src/buildroot
COPY tool-config /usr/src/buildroot/.config
COPY busybox.config /usr/src/buildroot/package/busybox/busybox.config
RUN make BR2_TOOLCHAIN_HEADERS_AT_LEAST="3.10" oldconfig
RUN make BR2_TOOLCHAIN_HEADERS_AT_LEAST="3.10" source
RUN make BR2_TOOLCHAIN_HEADERS_AT_LEAST="3.10" toolchain

COPY config /usr/src/buildroot/.config
RUN make BR2_TOOLCHAIN_HEADERS_AT_LEAST="3.10" oldconfig
RUN make BR2_TOOLCHAIN_HEADERS_AT_LEAST="3.10"
