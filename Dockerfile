# Build Frontend
FROM node:25-alpine3.22 AS frontend

WORKDIR /home/container
COPY ./frontend .

RUN npm install && npm run build

# Build Backend
FROM golang:1.25.4-alpine3.22 AS backend

WORKDIR /home/container

RUN apk add gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

ARG VERSION
ENV VERSION=$VERSION
ENV CGO_ENABLED=1

COPY . .

# Copy the built frontend only, need to find a better way to do this
RUN rm -rf ./frontend/*
COPY --from=frontend /home/container/build ./frontend/build
COPY ./frontend/embed.go ./frontend

RUN go build -v -ldflags "-X github.com/tubblelive/star.Version=$VERSION" -o /home/container/output cmd/main.go

# Runtime
FROM alpine:3.22.2 AS runtime

LABEL author=Santio
WORKDIR /home/container

RUN apk add sqlite-libs

COPY --from=backend /home/container/output ./star
RUN chmod +x ./star

CMD ["/home/container/star"]