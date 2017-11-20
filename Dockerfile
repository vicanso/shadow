FROM alpine

EXPOSE 3015 

ADD ./shadow /

CMD ["/shadow"]
