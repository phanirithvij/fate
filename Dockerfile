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
RUN apk add --no-cache build-base
RUN apk add --no-cache libc6-compat

# RUN mkdir -p /app
# RUN git clone https://github.com/phanirithvij/filebrowser.git /app/filebrowser
WORKDIR /app/filebrowser
RUN sh wizard.sh -d -c
RUN mv filebrowser /app/filebrowser-custom
WORKDIR /app
RUN go build --ldflags '-linkmode external -extldflags "-static"'
# RUN go build -tags netgo -a -v
# RUN go build

# Prepare final, minimal image
FROM heroku/heroku:18

COPY --from=build /app /app
ENV HOME /app
WORKDIR /app
RUN useradd -m heroku
USER heroku
ENV PATH="/app:${PATH}"
RUN ls -lsh
CMD /app/fate
