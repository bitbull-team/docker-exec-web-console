FROM alpine:latest

RUN apk --update add socat

ADD main /
ADD index.html /
ADD xterm /xterm

COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 8888

CMD ["/main"]
