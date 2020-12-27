FROM golang:alpine

CMD ["websocket-proxy"]
WORKDIR /go/src/websocket-proxy
EXPOSE 8783

COPY *.go ./
RUN go install
RUN rm *
