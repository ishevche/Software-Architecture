FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY logging_service.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /logging_service

FROM gcr.io/distroless/base-debian11 AS build-release-stage

COPY --from=build-stage /logging_service /logging_service

EXPOSE 8081

ENTRYPOINT ["/logging_service"]
