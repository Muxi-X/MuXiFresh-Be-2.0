FROM ubuntu:22.04

WORKDIR /MuxiFresh-Be-2.0/intro

COPY rpc/intro intro-rpc
COPY rpc/etc/intro.yaml etc/intro-rpc.yaml

COPY api/intro intro-api
COPY api/etc/intro-api.yaml etc/intro-api.yaml

COPY start.sh ./

RUN chmod +x ./start.sh

CMD ["./start.sh"]