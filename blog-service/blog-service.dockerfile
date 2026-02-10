#build a tiny docker image

FROM alpine:latest

RUN mkdir /app

COPY blogService /app

CMD ["/app/blogService"]
