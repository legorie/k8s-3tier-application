FROM golang:alpine AS build-env
RUN mkdir /go/src/web && apk update && apk add git
ADD . /go/src/web/
WORKDIR /go/src/web
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o web .

FROM scratch
WORKDIR /web
COPY --from=build-env /go/src/web .
ENTRYPOINT [ "./web" ]
EXPOSE 8080