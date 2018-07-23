FROM golang:1 AS builder

ENV REPO=github.com/xperimental/ipromnb
ENV PACKAGE=${REPO}/cmd/prometheus-kernel

RUN mkdir -p /go/src/${REPO}
WORKDIR /go/src/${REPO}

ENV LD_FLAGS="-w"

RUN apt-get update
RUN apt-get install -y libzmq3-dev

COPY . /go/src/${REPO}
WORKDIR /go/src/${PACKAGE}
RUN go install -a -v -tags netgo -ldflags "${LD_FLAGS}" .

FROM jupyter/base-notebook
LABEL maintainer="Robert Jacob <xperimental@solidproject.de>"

USER root

RUN apt-get update \
 && apt-get install -y libzmq3-dev \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/bin/prometheus-kernel /usr/local/bin/
COPY kernel.json logo-32x32.png logo-64x64.png /opt/conda/share/jupyter/kernels/prometheus/

VOLUME /home/jovyan/work/

USER $NB_UID
