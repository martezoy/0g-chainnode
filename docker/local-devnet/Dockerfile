FROM ubuntu:24.04

RUN apt-get update && apt-get install -y openssh-server git wget jq make gcc sudo
COPY 0gchaind /usr/local/bin/0gchaind
RUN chmod +x /usr/local/bin/0gchaind

COPY init-genesis.sh /root/init-genesis.sh
RUN chmod +x /root/init-genesis.sh
WORKDIR /root
RUN /bin/bash ./init-genesis.sh 192.168.1.3,192.168.1.4,192.168.1.5,192.168.1.6 12345678

CMD ["sleep", "infinity"]