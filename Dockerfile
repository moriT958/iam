# Build the Astro site (web/dist).
FROM node:24.22.1-alpine AS web
WORKDIR /app/web

RUN npm install -g pnpm@10

COPY web/pnpm-workspace.yaml web/package.json web/pnpm-lock.yaml ./
COPY web/packages/font-subset/package.json packages/font-subset/package.json
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

# Build the Go binary with web/dist embedded.
FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY web/embed.go ./web/embed.go
COPY --from=web /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/iam ./cmd/iam
# Empty directory for the SQLite database, owned by the nonroot user below.
RUN mkdir -p /out/data

# Run
FROM gcr.io/distroless/static:nonroot AS runtime
COPY --from=build /out/iam /iam
COPY --from=build --chown=65532:65532 /out/data /data
ENV IAM_DATA_DIR=/data
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/iam"]
