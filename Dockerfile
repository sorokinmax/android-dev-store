##
## Build stage
##
FROM golang:1.21.7-alpine3.19 AS build
RUN apk update && apk upgrade&& \
     apk add --no-cache git gcc g++ musl-dev
COPY . /src
WORKDIR /src

RUN GOOS=linux CGO_ENABLED=1 go build -o /astore ./cmd/astore/

##
## Final image stage
##
FROM alpine:3.19
ENV TZ=Etc/UTC
RUN apk add tzdata && cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && apk del tzdata
WORKDIR /astore
COPY --from=build /astore /astore/astore
COPY --from=build /src/cmd/astore/views /astore/views
COPY --from=build /src/cmd/astore/data /astore/data
RUN mkdir -p /astore/apps && mkdir -p /astore/tmp

EXPOSE 80
ENTRYPOINT ["/astore/astore"]