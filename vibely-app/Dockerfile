FROM node:22-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
ARG EXPO_PUBLIC_API_URL=http://localhost:7700/api/v1
ARG EXPO_PUBLIC_WS_URL=ws://localhost:7700/ws
ARG EXPO_PUBLIC_ALLOW_INSECURE_API=true
ENV EXPO_PUBLIC_API_URL=${EXPO_PUBLIC_API_URL}
ENV EXPO_PUBLIC_WS_URL=${EXPO_PUBLIC_WS_URL}
ENV EXPO_PUBLIC_ALLOW_INSECURE_API=${EXPO_PUBLIC_ALLOW_INSECURE_API}
RUN npx expo export --platform web

FROM nginx:1.30.6-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
