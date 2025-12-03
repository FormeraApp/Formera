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

# Environment variables (override with NUXT_PUBLIC_BASE_URL and NUXT_PUBLIC_API_URL)
ENV NUXT_PUBLIC_BASE_URL=http://localhost:3000
ENV NUXT_PUBLIC_API_URL=http://localhost:8080/api
ENV NITRO_PORT=3000

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
