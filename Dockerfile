# ---- Stage 1: Build ----
FROM golang:1.25.1 AS builder

WORKDIR /app

# Copiar go.mod y go.sum primero (mejor cacheo de dependencias)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código
COPY . .

# Compilar el binario
RUN CGO_ENABLED=0 GOOS=linux go build -o orquestador-notificacion ./cmd/orchestrator

# ---- Stage 2: Runtime ----
FROM alpine:3.20

WORKDIR /app

# Copiar el binario compilado
COPY --from=builder /app/orquestador-notificacion .

EXPOSE 8080

CMD ["./orquestador-notificacion"]