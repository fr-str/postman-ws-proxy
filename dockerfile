FROM golang:1.19.0 as build

RUN mkdir /build
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build 


RUN mkdir /app
WORKDIR /app

FROM alpine:3.16.2

RUN mkdir /app
WORKDIR /app
COPY --from=build /build/postman-proxy pp
ENV PP_PORT=8008 \
    PP_LOG_FILE_PATH=/app/log-files \
    PP_LOG_LEVEL=1

CMD ["/bin/sh","-c","/app/pp"]


