FROM sourcegraph/alpine-3.14:server

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN apk add --no-cache \
    bash

RUN mkdir -p /data && chown -R sourcegraph:sourcegraph /data
USER sourcegraph

EXPOSE 9000
WORKDIR /
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/blobstore"]
COPY blobstore /usr/local/bin/
