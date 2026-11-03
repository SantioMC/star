FROM oven/bun:1-alpine AS frontend
WORKDIR /app/frontend

COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile --ignore-scripts

COPY frontend/ ./
RUN bun run build

FROM golang:1.26-alpine AS backend
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY emails/ emails/
COPY internal/ internal/
COPY migrations/ migrations/
COPY frontend/embed.go frontend/embed.go
COPY --from=frontend /app/frontend/build frontend/build

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /star ./cmd

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=backend /star /star

EXPOSE 8080

ENTRYPOINT ["/star"]
