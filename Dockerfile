#
# release container
#
FROM alpine:latest
RUN apk --no-cache add \
    ca-certificates \
    bash \
    git

WORKDIR /bin/
COPY ./heroku-listing ./heroku-listing

ENTRYPOINT [ "/bin/heroku-listing" ]
