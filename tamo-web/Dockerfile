FROM node:24-alpine AS build
ARG VERSION=dev
ARG SOURCE_REVISION=unknown
ARG NEXT_PUBLIC_IDENTITY_API=http://localhost:4000/api/v1
ARG NEXT_PUBLIC_MEDIA_API=http://localhost:5000/api/v1
ARG NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN=http://localhost:9000
ARG API_TIMEOUT_MS=5000
ARG APP_ORIGIN=http://localhost:3000
ENV NEXT_PUBLIC_IDENTITY_API=$NEXT_PUBLIC_IDENTITY_API \
    NEXT_PUBLIC_MEDIA_API=$NEXT_PUBLIC_MEDIA_API \
    NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN=$NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN \
    API_TIMEOUT_MS=$API_TIMEOUT_MS \
    APP_ORIGIN=$APP_ORIGIN
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
LABEL org.opencontainers.image.title="Tamo Images Web" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$SOURCE_REVISION
COPY --from=build --chown=node:node /app/.next/standalone ./
COPY --from=build --chown=node:node /app/.next/static ./.next/static
EXPOSE 3000
USER node
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget --quiet --tries=1 --spider http://127.0.0.1:3000/ || exit 1
CMD ["node", "server.js"]
