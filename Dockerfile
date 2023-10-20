FROM golang:1.21.3-alpine AS build
WORKDIR $GOPATH/src/github.com/micromata/dave/
COPY . .
RUN go build -o /go/bin/dave cmd/dave/main.go
RUN go build -o /go/bin/davecli cmd/davecli/main.go

FROM alpine:latest  
RUN addgroup -g 1000 dave
RUN adduser -S -G dave -u 1000 dave
COPY --from=build /go/bin/davecli /usr/local/bin
COPY --from=build /go/bin/dave /usr/local/bin
USER dave
ENTRYPOINT ["/usr/local/bin/dave"]
