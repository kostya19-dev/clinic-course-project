FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/clinic-api ./cmd/api

FROM alpine:3.21

WORKDIR /app
COPY --from=build /out/clinic-api /app/clinic-api

EXPOSE 8080

ENTRYPOINT ["/app/clinic-api"]
