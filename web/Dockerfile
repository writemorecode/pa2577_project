FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /web

FROM gcr.io/distroless/base-debian11 as build-release-stage

WORKDIR /

COPY --from=build-stage /web /web
COPY *.html ./
COPY *.js ./

EXPOSE 8080

ENTRYPOINT ["/web"]

