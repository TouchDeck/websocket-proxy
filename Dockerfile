FROM golang:alpine

CMD ["tcp-proxy"]
WORKDIR /go/src/tcp-proxy
EXPOSE 6061 6062

COPY *.go ./
RUN go install
RUN rm *
