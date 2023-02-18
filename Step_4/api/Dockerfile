FROM golang:alpine AS build-env
RUN mkdir /go/src/api && apk update && apk add git
ADD . /go/src/api/
WORKDIR /go/src/api
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o api .

FROM scratch
WORKDIR /api
COPY --from=build-env /go/src/api .
ENTRYPOINT [ "./api" ]
EXPOSE 8090