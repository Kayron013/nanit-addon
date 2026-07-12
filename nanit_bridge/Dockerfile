# BUILD_FROM must be declared before the first FROM to be visible in any
# FROM line (an ARG between stages is scoped to the preceding stage).
ARG BUILD_FROM

# Stage 1: frontend (Next.js static export)
FROM node:18-alpine AS frontend-build
WORKDIR /build/frontend
COPY nanit-src/frontend/package.json nanit-src/frontend/package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi
COPY nanit-src/frontend/ ./
RUN npm run build

# Stage 2: backend (Go, CGO for sqlite — built natively per-arch by the Supervisor)
FROM golang:1.24-bookworm AS backend-build
WORKDIR /build
COPY nanit-src/go.mod nanit-src/go.sum ./
RUN go mod download
COPY nanit-src/cmd ./cmd
COPY nanit-src/pkg ./pkg
COPY nanit-src/scripts ./scripts
COPY --from=frontend-build /build/frontend/dist ./web
RUN CGO_ENABLED=1 go build -ldflags "-X main.GitCommit=addon-88f3d37" \
    -o ./bin/nanit ./cmd/nanit/*.go

# Stage 3: runtime on Home Assistant base image (provides bashio)
FROM ${BUILD_FROM}

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates ffmpeg sqlite3 libsqlite3-0 \
    && apt-get autoremove -y && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend-build /build/bin/nanit /app/bin/nanit
COPY --from=backend-build /build/scripts /app/scripts
COPY --from=backend-build /build/web /app/web
RUN chmod +x /app/scripts/*.sh

COPY run.sh /
RUN chmod a+x /run.sh

WORKDIR /app
CMD [ "/run.sh" ]
