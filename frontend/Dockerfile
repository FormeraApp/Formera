# Build Frontend (Nuxt)
FROM node:20-alpine AS builder

RUN corepack enable && corepack prepare yarn@4.9.4 --activate

WORKDIR /app

COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable || yarn install

COPY . .

RUN yarn build

# Final Image
FROM node:20-alpine

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/.output ./.output
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

# Environment variables - use simple BASE_URL and API_URL
ENV BASE_URL=http://localhost:3000
ENV API_URL=http://localhost:8080
ENV NITRO_PORT=3000

EXPOSE 3000

RUN addgroup -S app && adduser -S app -G app \
  && chown -R app:app /app
USER app
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["node", ".output/server/index.mjs"]
