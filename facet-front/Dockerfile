FROM node:22.20.0-alpine AS builder

WORKDIR /app

COPY package*.json ./

RUN npm ci

COPY . .

RUN npm run build

FROM nginx:1.27.0-alpine

RUN adduser -D -H -u 1001 -s /bin/bash appuser

COPY --from=builder /app/dist /usr/share/nginx/html

COPY --from=builder /app/website /usr/share/nginx/html/website

COPY nginx.conf /etc/nginx/conf.d/default.conf

RUN chown -R appuser:appuser /usr/share/nginx/html /var/cache/nginx /var/run

EXPOSE 80

USER appuser

CMD ["nginx", "-g", "daemon off;"]
