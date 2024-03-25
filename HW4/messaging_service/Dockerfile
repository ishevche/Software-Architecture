FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY messaging_service.go ./

RUN go build -o /messaging_service

FROM golang:1.22 AS build-release-stage

COPY --from=build-stage /messaging_service /messaging_service

EXPOSE 8082

ENTRYPOINT ["/messaging_service"]
