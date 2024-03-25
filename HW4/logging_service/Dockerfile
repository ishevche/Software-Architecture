FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY logging_service.go ./

RUN go build -o /logging_service

FROM golang:1.22 AS build-release-stage

COPY --from=build-stage /logging_service /logging_service

EXPOSE 8081

ENTRYPOINT ["/logging_service"]
