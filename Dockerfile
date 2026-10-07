FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/fuzeflow ./cmd/server

FROM alpine:3.22
RUN addgroup -S fuzeflow && adduser -S -G fuzeflow fuzeflow
COPY --from=build /out/fuzeflow /usr/local/bin/fuzeflow
USER fuzeflow
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/fuzeflow"]
