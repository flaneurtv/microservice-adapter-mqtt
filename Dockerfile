FROM node:onbuild

RUN apt-get update && \
  apt-get install -y jq uuid-runtime && \
  apt-get autoclean -y && apt-get autoremove -y && apt-get clean -y && \
  rm -rf /var/lib/apt/lists/*
