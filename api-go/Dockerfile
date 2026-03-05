FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=build /bin/api /bin/api
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
