FROM node:25-bookworm-slim@sha256:32f45869cf02c26971de72c383d5f99cab002905ed8b515b56df925007941782 AS base
# FROM gcr.io/distroless/nodejs22-debian12 AS base
WORKDIR /src

# Create non-root user early (better layer caching)
RUN groupadd -r nodejs && useradd -r -g nodejs nodejs

WORKDIR /app

# ---- Dependencies Layer ----
FROM base AS deps

COPY package.json package-lock.json ./

# Deterministic install, no dev deps
RUN npm ci --omit=dev

# ---- Runtime Layer ----
FROM base AS runtime

ENV NODE_ENV=production

WORKDIR /app

# Copy node_modules from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy only required runtime files
COPY src ./src
COPY migrations ./migrations
COPY package.json ./

# Change ownership to non-root
RUN chown -R nodejs:nodejs /app

USER nodejs

CMD ["node", "src/server.js"]
