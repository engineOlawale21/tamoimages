FROM node:24-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build
FROM node:24-alpine
WORKDIR /app
ENV NODE_ENV=production
ARG VERSION=dev
ARG SOURCE_REVISION=unknown
LABEL org.opencontainers.image.title="Tamo Identity Service" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$SOURCE_REVISION
COPY package*.json ./
RUN npm ci --omit=dev
COPY --from=build /app/dist ./dist
COPY migrations ./migrations
EXPOSE 4000
USER node
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget --quiet --tries=1 --spider http://127.0.0.1:4000/api/v1/health/live || exit 1
CMD ["node", "dist/main.js"]
