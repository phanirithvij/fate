FROM node:14
COPY . /app
WORKDIR /app

RUN mkdir -p /tmp/filebrowser
RUN git clone https://github.com/phanirithvij/filebrowser.git /tmp/filebrowser/filebrowser
WORKDIR /tmp/filebrowser/filebrowser
RUN bash wizard.sh -d -a
RUN pwd && ls -lSh * && tree
COPY frontend/dist /app

FROM heroku/heroku:18-build as build
COPY . /app
WORKDIR /app

# Setup buildpack
RUN mkdir -p /tmp/buildpack/heroku/go /tmp/build_cache /tmp/env
RUN curl https://codon-buildpacks.s3.amazonaws.com/buildpacks/heroku/go.tgz | tar xz -C /tmp/buildpack/heroku/go

RUN mkdir -p /tmp/filebrowser
RUN git clone https://github.com/phanirithvij/filebrowser.git /tmp/filebrowser/filebrowser
WORKDIR /tmp/filebrowser/filebrowser
RUN bash wizard.sh -d -b
RUN pwd && ls -lSh && tree
RUN mv filebrowser filebrowser-custom
COPY filebrowser-custom /app

#Execute Buildpack
RUN STACK=heroku-18 /tmp/buildpack/heroku/go/bin/compile /app /tmp/build_cache /tmp/env

# Prepare final, minimal image
FROM heroku/heroku:18

COPY --from=build /app /app
ENV HOME /app
WORKDIR /app
RUN useradd -m heroku
USER heroku
CMD /app/fate
