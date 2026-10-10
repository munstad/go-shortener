FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /shortener .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /shortener /usr/local/bin/shortener
VOLUME /app/data
EXPOSE 8080
CMD ["shortener", "-data", "/app/data/links.json"]
