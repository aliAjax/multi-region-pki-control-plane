FROM golang:1.23 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/pki-api ./cmd/pki-api
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/pki-api /pki-api
EXPOSE 8080
ENTRYPOINT ["/pki-api"]
