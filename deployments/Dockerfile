FROM golang:1.23.2-bullseye as build
RUN apt-get update \
    && apt-get install -y --force-yes --no-install-recommends apt-transport-https curl ca-certificates \
    && apt-get clean \
    && apt-get autoremove \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY . .
RUN go build -v -o microsaas cmd/app.go

FROM debian:11.8-slim
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=build /app/microsaas /usr/local/bin/microsaas
ENV GIN_MODE=release
CMD [ "/usr/local/bin/microsaas", "--migrate" ] 
EXPOSE 80