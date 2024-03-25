FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY facade_service.go ./

RUN go build -o /facade_service

FROM golang:1.22 AS build-release-stage

COPY --from=build-stage /facade_service /facade_service

EXPOSE 8080

ENTRYPOINT ["/facade_service"]
