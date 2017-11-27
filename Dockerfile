FROM ubuntu 

EXPOSE 3015 

RUN apt-get update \
  && apt-get install -y ca-certificates

ADD ./shadow /

CMD ["/shadow"]
