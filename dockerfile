FROM golang:1.19.0 as build


RUN mkdir /app
WORKDIR /app
COPY . .
RUN go build 

ENV PP_PORT=8008

CMD ["/bin/sh","-c","/app/postman-proxy"]


