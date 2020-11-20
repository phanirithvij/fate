FROM node:14 as node
COPY . /app
WORKDIR /app

RUN mkdir -p /tmp/filebrowser
RUN git clone https://github.com/phanirithvij/filebrowser.git /tmp/filebrowser/filebrowser
WORKDIR /tmp/filebrowser/filebrowser
RUN sh wizard.sh -d -a

FROM golang:1.15.5-alpine AS build
COPY --from=node /app /app
RUN apk update && apk install git
# COPY --from=node /tmp/filebrowser /tmp/filebrowser

RUN mkdir -p /tmp/filebrowser
RUN git clone https://github.com/phanirithvij/filebrowser.git /tmp/filebrowser/filebrowser
WORKDIR /tmp/filebrowser/filebrowser
RUN sh wizard.sh -d -c
RUN mv filebrowser /app/filebrowser-custom

# Prepare final, minimal image
FROM heroku/heroku:18

COPY --from=build /app /app
ENV HOME /app
WORKDIR /app
RUN useradd -m heroku
USER heroku
CMD /app/fate
