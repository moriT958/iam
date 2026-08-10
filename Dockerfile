FROM node:22-alpine AS build
WORKDIR /app

RUN npm install -g pnpm@10

COPY pnpm-workspace.yaml package.json pnpm-lock.yaml ./
COPY packages/font-subset/package.json packages/font-subset/package.json
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

FROM nginx:alpine AS runtime
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
