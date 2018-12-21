#
# release container
#
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /bin/
COPY ./redis-sl-fwd-to-sumologic ./redis-sl-fwd-to-sumologic

EXPOSE     9121
ENTRYPOINT [ "/bin/redis-sl-fwd-to-sumologic" ]
