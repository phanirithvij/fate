FROM node:14 as node
COPY . /app
WORKDIR /app

RUN mkdir -p /app/filebrowser
RUN git clone https://github.com/phanirithvij/filebrowser.git /app/filebrowser
WORKDIR /app/filebrowser
RUN sh wizard.sh -d -a
RUN rm -rf /app/filebrowser/frontend/node_modules

FROM golang:1.15.5-alpine AS build
COPY --from=node /app /app
RUN apk update
RUN apk add git gcc musl-dev
# RUN mkdir -p /app
# RUN git clone https://github.com/phanirithvij/filebrowser.git /app/filebrowser
WORKDIR /app/filebrowser
RUN sh wizard.sh -d -c
RUN mv filebrowser /app/filebrowser-custom

# Prepare final, minimal image
FROM heroku/heroku:18

COPY --from=build /app /app
ENV HOME /app
WORKDIR /app
RUN useradd -m heroku
USER heroku
RUN ls -lsh
RUN go build
CMD /app/fate
