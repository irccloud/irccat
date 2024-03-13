FROM debian:stable-slim AS build
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY irccat /

EXPOSE 12345
EXPOSE 8045

CMD ["/irccat"]
