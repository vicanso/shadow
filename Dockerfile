FROM alpine

EXPOSE 3015 

RUN apk --no-cache add ca-certificates \
  && update-ca-certificates

ADD ./shadow /

CMD ["/shadow"]
