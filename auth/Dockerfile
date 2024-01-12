FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth

#FROM build-stage AS run-test-stage
# Insert "go test ..." here

FROM gcr.io/distroless/base-debian11 as build-release-stage

WORKDIR /

COPY --from=build-stage /auth /auth

EXPOSE 8080

ENTRYPOINT ["/auth"]

