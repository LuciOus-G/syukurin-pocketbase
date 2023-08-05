FROM golang:alpine

WORKDIR /pocketbase
COPY . /pocketbase

RUN cd /pocketbase/app/base \
    && go build -o ../../server

EXPOSE 80
EXPOSE 443

# start PocketBase
CMD ["/pocketbase/server", "serve", "--http=0.0.0.0:80"]

