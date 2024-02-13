FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY facade_service.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /facade_service

FROM gcr.io/distroless/base-debian11 AS build-release-stage

COPY --from=build-stage /facade_service /facade_service

EXPOSE 8080

ENTRYPOINT ["/facade_service"]
