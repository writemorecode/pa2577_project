FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /images

#FROM build-stage AS run-test-stage
# Insert "go test ..." here

FROM gcr.io/distroless/base-debian11 as build-release-stage

WORKDIR /

COPY --from=build-stage /images /images

EXPOSE 8080

ENTRYPOINT ["/images"]

