FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ingestion-service .

FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=build /out/ingestion-service /ingestion-service
EXPOSE 8081
ENTRYPOINT ["/ingestion-service"]
