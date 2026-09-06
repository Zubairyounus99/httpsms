# Stage 1: Build the Nuxt application
FROM node:22-alpine AS build

WORKDIR /app

# Copy dependency definitions
COPY package.json pnpm-lock.yaml ./

# Install pnpm v9 (avoids pnpm v10 ERR_PNPM_IGNORED_BUILDS)
RUN npm install -g pnpm@9

# Install dependencies ignoring git/prepare scripts (skips husky error)
RUN pnpm install --frozen-lockfile --ignore-scripts

# Copy project files
COPY . .

# Generate static assets
RUN pnpm run generate

# Stage 2: Serve the application with Nginx
FROM nginx:stable-alpine AS production

# Copy built files from previous stage
COPY --from=build /app/.output/public /usr/share/nginx/html

# Copy the nginx configuration file
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
