FROM golang:1.11-alpine
RUN apk add --no-cache git
RUN go get -u github.com/micromata/dave/cmd/...

FROM alpine:latest  
RUN apk update && apk upgrade
RUN adduser -S dave
COPY --from=0 /go/bin/dave /usr/local/bin
USER dave
ENTRYPOINT ["/usr/local/bin/dave"]
