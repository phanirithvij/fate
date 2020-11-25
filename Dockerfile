FROM node:14 as node
COPY . /app
WORKDIR /app

# build assets only we don't need filebrowser exe
RUN bash custom-fb.sh -a -d
# remove node_modules as we need not copy it
RUN rm -rf /app/filebrowser/frontend/node_modules

FROM golang:1.15.5-alpine AS build
COPY --from=node /app /app
RUN apk update
# https://stackoverflow.com/a/58478169/8608146
RUN apk add git gcc build-base libc6-compat
# RUN apk add git gcc musl-dev build-base libc6-compat

WORKDIR /app
# rice assets here where we have go available
RUN sh custom-fb.sh -r -d
ENV GLIBC_REPO=https://github.com/sgerrand/alpine-pkg-glibc
ENV GLIBC_VERSION=2.32-r0

RUN set -ex && \
    apk --update add libstdc++ curl ca-certificates && \
    for pkg in glibc-${GLIBC_VERSION} glibc-bin-${GLIBC_VERSION}; \
        do curl -sSL ${GLIBC_REPO}/releases/download/${GLIBC_VERSION}/${pkg}.apk -o /tmp/${pkg}.apk; done && \
    apk add --allow-untrusted /tmp/*.apk && \
    rm -v /tmp/*.apk && \
    /usr/glibc-compat/sbin/ldconfig /lib /usr/glibc-compat/lib

RUN go build
# RUN go build --ldflags '-linkmode external'
RUN ls -lsh
RUN file fate

# Prepare final, minimal image
FROM heroku/heroku:18

COPY --from=build /app /app
ENV HOME /app
WORKDIR /app
RUN rm /bin/sh && ln -s /bin/bash /bin/sh
ADD ./.profile.d /app/.profile.d
RUN useradd -m heroku
USER heroku
# https://stackoverflow.com/a/38742545/8608146
ENV PATH="/app:${PATH}"
RUN ls -lsh
CMD pwd && ls -lsh && sleep 100000
