FROM sourcegraph/alpine-3.14:server

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN apk update && apk add --no-cache \
    tini

USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/migrator"]
COPY migrator /usr/local/bin/
COPY schema-descriptions /schema-descriptions
