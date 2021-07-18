FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o tailscale_http_sd

FROM alpine
WORKDIR /app
COPY --from=build-env /src/tailscale_http_sd /app/
ENTRYPOINT ["/app/tailscale_http_sd"]
