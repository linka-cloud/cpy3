FROM python:3.8

RUN curl -sL https://go.dev/dl/go1.20.3.linux-amd64.tar.gz | tar xvz -C /usr/local/

ENV PATH="/usr/local/go/bin:${PATH}"

ENV PKG_CONFIG_PATH="/usr/local/lib/pkgconfig"
