# Copyright IBM Corp. 2015, 2025
# SPDX-License-Identifier: MPL-2.0

FROM debian:stable

RUN apt-get update -y && apt-get install --no-install-recommends -y -q \
                         curl \
                         zip \
                         build-essential \
                         ca-certificates \
                         git mercurial bzr \
               && rm -rf /var/lib/apt/lists/*

ENV GOVERSION 1.9.2
RUN mkdir /goroot && mkdir /gopath
RUN curl https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz \
           | tar xvzf - -C /goroot --strip-components=1

ENV GOPATH /gopath
ENV GOROOT /goroot
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH

RUN go get github.com/mitchellh/gox

RUN mkdir -p /gopath/src/github.com/hashicorp/vault-ssh-helper
WORKDIR /gopath/src/github.com/hashicorp/vault-ssh-helper
CMD make bin
