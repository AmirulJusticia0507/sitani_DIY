FROM node:22.23.4-alpine AS dependencies

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:22.23.4-alpine AS builder

WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=dependencies /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:22.23.4-alpine AS runner

WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME=0.0.0.0
RUN addgroup -S sitani && adduser -S -G sitani sitani
COPY --from=builder --chown=sitani:sitani /app/public ./public
COPY --from=builder --chown=sitani:sitani /app/.next/standalone ./
COPY --from=builder --chown=sitani:sitani /app/.next/static ./.next/static

USER sitani
EXPOSE 3000
CMD ["node", "server.js"]
