FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY messaging_service.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /messaging_service

FROM gcr.io/distroless/base-debian11 AS build-release-stage

COPY --from=build-stage /messaging_service /messaging_service

EXPOSE 8082

ENTRYPOINT ["/messaging_service"]
