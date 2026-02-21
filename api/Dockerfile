FROM node:25-alpine3@sha256:b9b5737eabd423ba73b21fe2e82332c0656d571daf1ebf19b0f89d0dd0d3ca93 AS base

RUN apk update && apk upgrade --no-cache
RUN npm install -g npm@latest

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

USER node

CMD ["node", "src/server.js"]