FROM node:25-alpine3@sha256:b9b5737eabd423ba73b21fe2e82332c0656d571daf1ebf19b0f89d0dd0d3ca93 AS base

RUN apk update && apk upgrade --no-cache
RUN npm install -g npm@latest

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --omit=dev

# ---------- Runtime ----------
FROM gcr.io/distroless/nodejs24-debian13 AS runtime

WORKDIR /app

# copy production node_modules
COPY --from=base /app/node_modules ./node_modules

# copy application source
COPY src ./src
COPY migrations ./migrations
COPY package.json ./

USER nonroot
CMD ["src/server.js"]

# ---------- Migration Image ----------
FROM base AS migrate
WORKDIR /app

COPY --from=base /app/node_modules ./node_modules
COPY src ./src
COPY migrations ./migrations
COPY package.json ./

CMD ["node", "migrations/cli.js"]