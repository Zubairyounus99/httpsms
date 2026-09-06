# Stage 1: Build the Nuxt application
FROM node:22-alpine AS build

WORKDIR /app

# Build arguments for frontend static compilation
ARG NUXT_PUBLIC_API_BASE_URL
ARG NUXT_PUBLIC_FIREBASE_API_KEY
ARG NUXT_PUBLIC_FIREBASE_AUTH_DOMAIN
ARG NUXT_PUBLIC_FIREBASE_PROJECT_ID
ARG NUXT_PUBLIC_FIREBASE_STORAGE_BUCKET
ARG NUXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID
ARG NUXT_PUBLIC_FIREBASE_APP_ID
ARG NUXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY

# Convert build arguments to environment variables for Nuxt generator
ENV NUXT_PUBLIC_API_BASE_URL=$NUXT_PUBLIC_API_BASE_URL
ENV NUXT_PUBLIC_FIREBASE_API_KEY=$NUXT_PUBLIC_FIREBASE_API_KEY
ENV NUXT_PUBLIC_FIREBASE_AUTH_DOMAIN=$NUXT_PUBLIC_FIREBASE_AUTH_DOMAIN
ENV NUXT_PUBLIC_FIREBASE_PROJECT_ID=$NUXT_PUBLIC_FIREBASE_PROJECT_ID
ENV NUXT_PUBLIC_FIREBASE_STORAGE_BUCKET=$NUXT_PUBLIC_FIREBASE_STORAGE_BUCKET
ENV NUXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=$NUXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID
ENV NUXT_PUBLIC_FIREBASE_APP_ID=$NUXT_PUBLIC_FIREBASE_APP_ID
ENV NUXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY=$NUXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY

# Copy dependency files
COPY package.json pnpm-lock.yaml* ./

# Install pnpm v9
RUN npm install -g pnpm@9

# Install dependencies skipping husky/git hooks
RUN pnpm install --ignore-scripts

# Copy application source code
COPY . .

# Generate Nuxt types and compile static output directly with npx
RUN npx nuxt prepare && npx nuxt generate

# Stage 2: Serve the application with Nginx
FROM nginx:stable-alpine AS production

# Copy built files from previous stage
COPY --from=build /app/.output/public /usr/share/nginx/html

# Copy the nginx configuration file
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
