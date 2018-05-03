FROM alpine:3.7

RUN apk add --no-cache ca-certificates

ADD bin/dashboard-controller_linux_amd64 /usr/local/bin/dashboard-controller

ENTRYPOINT ["/usr/local/bin/dashboard-controller"]
ARG VCS_REF
LABEL org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/presslabs/dashboard.git" \
