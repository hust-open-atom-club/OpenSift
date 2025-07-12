# syntax=docker.io/docker/dockerfile:1

FROM node:22-bookworm-slim AS base

# Install dependencies only when needed
FROM base AS pub_deps

WORKDIR /app


# 设置 yarn 源为国内镜像，并增加超时时间
RUN yarn config set registry https://registry.npmmirror.com
RUN yarn config set network-timeout 600000 -g

# Install dependencies based on the preferred package manager
COPY public/package.json public/yarn.lock* public/package-lock.json* public/pnpm-lock.yaml* public/.npmrc* ./
RUN \
if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i --frozen-lockfile; \
  else echo "Lockfile not found." && exit 1; \
  fi

FROM base AS admin_deps
WORKDIR /app


# 设置 yarn 源为国内镜像，并增加超时时间
RUN yarn config set registry https://registry.npmmirror.com
RUN yarn config set network-timeout 600000 -g

# Install dependencies based on the preferred package manager
COPY admin/package.json admin/yarn.lock* admin/package-lock.json* admin/pnpm-lock.yaml* admin/.npmrc* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i --frozen-lockfile; \
  else echo "Lockfile not found." && exit 1; \
  fi


# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app/public
COPY --from=pub_deps /app/node_modules ./node_modules
COPY public/ ./

ENV NEXT_TELEMETRY_DISABLED=1
ARG BACKEND

RUN \
  export BACKEND=$BACKEND; \
  if [ -f yarn.lock ]; then yarn run build; \
  elif [ -f package-lock.json ]; then npm run build; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm run build; \
  else echo "Lockfile not found." && exit 1; \
  fi

WORKDIR /app/admin
COPY --from=admin_deps /app/node_modules ./node_modules
COPY admin/ ./

RUN \
  if [ -f yarn.lock ]; then yarn run build; \
  elif [ -f package-lock.json ]; then npm run build; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm run build; \
  else echo "Lockfile not found." && exit 1; \
  fi

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN apt-get update && apt-get install -y --no-install-recommends \
  ca-certificates nginx \
  && rm -rf /var/lib/apt/lists/*

RUN npm install -g pm2

COPY deploy/ /app/deploy

# copy public

COPY --from=builder /app/public/public /app/public/public
# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder /app/public/.next/standalone /app/public/
COPY --from=builder /app/public/.next/static /app/public/.next/static

# copy admin

COPY --from=builder /app/admin/dist/* /app/admin/

WORKDIR /app/

EXPOSE 80

CMD ["pm2-runtime", "/app/deploy/ecosystem.config.js"]