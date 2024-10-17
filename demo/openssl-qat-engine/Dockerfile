FROM ubuntu:24.04 AS builder

ARG QATLIB_VERSION="24.09.0"
ARG QAT_ENGINE_VERSION="v1.6.2"
ARG IPSEC_MB_VERSION="v1.5"
ARG IPP_CRYPTO_VERSION="cryptography-primitives_1_0_0"

RUN apt update && \
    env DEBIAN_FRONTEND=noninteractive apt install -y \
    git \
    autoconf \
    libtool \
    libssl-dev \
    libnuma-dev \
    nasm \
    pkg-config \
    make \
    zlib1g-dev \
    cmake \
    python3

RUN git clone --depth 1 -b $QAT_ENGINE_VERSION https://github.com/intel/QAT_Engine && \
    git clone --depth 1 -b $IPP_CRYPTO_VERSION https://github.com/intel/ipp-crypto && \
    git clone --depth 1 -b $IPSEC_MB_VERSION https://github.com/intel/intel-ipsec-mb && \
    git clone --depth 1 -b $QATLIB_VERSION https://github.com/intel/qatlib

RUN cd /qatlib && \
    ./autogen.sh && \
    ./configure \
    --prefix=/usr \
    --enable-systemd=no && \
    make -j && \
    make install samples-install

RUN cd /ipp-crypto/sources/ippcp/crypto_mb && \
    cmake . -B"../build" \
    -DOPENSSL_INCLUDE_DIR=/usr/include/openssl \
    -DOPENSSL_LIBRARIES=/usr/lib64 \
    -DOPENSSL_ROOT_DIR=/usr/bin/openssl && \
    cd ../build && \
    make crypto_mb && make install

RUN cd /intel-ipsec-mb && \
    make && make install LIB_INSTALL_DIR=/usr/lib64

RUN cd /QAT_Engine && \
    sed -i '/fprintf(stderr, "QAT_HW device not available & QAT_SW not enabled. Using OpenSSL_SW!\\n");/ s/$/ return ret;/' e_qat.c && \
    ./autogen.sh && \
    ./configure \
    --disable-qat_sw && \
    make && make test && make install


FROM ubuntu:24.04

COPY --from=builder /QAT_Engine/testapp /usr/bin/
COPY --from=builder /usr/bin/*_sample* /usr/bin/
COPY --from=builder /usr/lib64/libIPSec_MB.so.1 /usr/lib/x86_64-linux-gnu/
COPY --from=builder /usr/lib/x86_64-linux-gnu/engines-3/qatengine.so /usr/lib/x86_64-linux-gnu/engines-3/qatengine.so
COPY --from=builder /usr/share/qat/calgary32 /usr/share/qat/
COPY --from=builder /qatlib/LICENSE /usr/share/package-licenses/qatlib/LICENSE
COPY --from=builder /QAT_Engine/LICENSE /usr/share/package-licenses/QAT_Engine/LICENSE
COPY --from=builder /ipp-crypto/LICENSE /usr/share/package-licenses/ipp-crypto/LICENSE
COPY --from=builder /intel-ipsec-mb/LICENSE /usr/share/package-licenses/intel-ipsec-mb/LICENSE
RUN ldconfig && apt update && env DEBIAN_FRONTEND=noninteractive apt install -y openssl libcrypto-mb-dev libqat4 libusdm-dev
