FROM node:14 as node
COPY . /app
WORKDIR /app

# RUN mkdir -p /app/filebrowser
# RUN git clone https://github.com/phanirithvij/filebrowser.git /app/filebrowser
WORKDIR /app/filebrowser
# build assets only we don't need filebrowser exe
RUN ls -la
RUN sh wizard.sh -d -a
RUN rm -rf /app/filebrowser/frontend/node_modules

FROM golang:1.15.5-alpine AS build
COPY --from=node /app /app
RUN apk update
# https://stackoverflow.com/a/58478169/8608146
RUN apk add git gcc musl-dev
RUN apk add build-base
RUN apk add libc6-compat

WORKDIR /app
RUN go build

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
CMD /app/fate
