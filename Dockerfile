FROM golang:latest AS build
WORKDIR /go/src/websocket-proxy
COPY . .

ENV CGO_ENABLED=0
RUN go get -d
RUN go install

FROM scratch
COPY --from=build /go/bin/websocket-proxy /bin/websocket-proxy
EXPOSE 8783
CMD ["websocket-proxy"]
