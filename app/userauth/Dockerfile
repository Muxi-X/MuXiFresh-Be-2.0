FROM centos:8.1.1911

WORKDIR /MuxiFresh-Be-2.0/userauth

#rpc
COPY cmd/rpc/accountCenter/accountCenter .
COPY cmd/rpc/accountCenter/etc/accountCenter.yaml etc/accountCenter.yaml

#api
COPY cmd/api/user-auth .
COPY cmd/api/etc/user-auth.yaml etc/user-auth.yaml



COPY start.sh ./

RUN chmod +x ./start.sh

CMD ["./start.sh"]