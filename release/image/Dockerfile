FROM debian:10.9

RUN rm -rf /var/lib/apt/lists/* && apt-get update -y \
    && apt-get install -y git make

COPY release.sh /

WORKDIR /operator-sdk

ENTRYPOINT ["/release.sh"]
