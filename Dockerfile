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
RUN apk add git gcc musl-dev build-base libc6-compat

WORKDIR /app
# rice assets here where we have go available
RUN sh custom-fb.sh -r -d
RUN go build --ldflags '-linkmode external'
RUN ls -lsh

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
