# build
FROM golang:1.26-alpine AS builder

# pasta em cmd/ a compilar: service-a ou service-b
ARG SERVICE
RUN test -n "$SERVICE" || (echo "build arg SERVICE é obrigatório" && exit 1)

RUN apk add --no-cache ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

# imagem final: só o binário + certificados TLS (ViaCEP e WeatherAPI são HTTPS)
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/app /app

# usuário sem privilégios (nobody)
USER 65534:65534

ENTRYPOINT ["/app"]
