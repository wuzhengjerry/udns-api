FROM mirrors.tencent.com/iyunwei/centos-l5_teg-agent-v1.0.1:7.2
EXPOSE 8080
WORKDIR /data/udns

#COPY ./supervisord.conf /etc/supervisord.conf
COPY ./docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

COPY udns-api /data/udns/
ENTRYPOINT ["/docker-entrypoint.sh"]