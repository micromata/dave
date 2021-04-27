FROM golang:1.11-alpine AS build
COPY . src/github.com/micromata/dave/
RUN go build -o /go/bin/dave /go/src/github.com/micromata/dave/cmd/dave/main.go
RUN go build -o /go/bin/davecli /go/src/github.com/micromata/dave/cmd/davecli/main.go

FROM alpine:latest  
RUN adduser -S dave
COPY --from=build /go/bin/davecli /usr/local/bin
COPY --from=build /go/bin/dave /usr/local/bin
USER dave
ENTRYPOINT ["/usr/local/bin/dave"]
