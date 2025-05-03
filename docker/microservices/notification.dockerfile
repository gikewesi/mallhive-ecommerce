# Stage 1: Builder
FROM node:18-alpine AS builder

WORKDIR /app

# Copy only package files and install full deps
COPY package*.json ./
RUN npm install

# Copy the entire app and build it
COPY . .
RUN npm run build


# Stage 2: Production image
FROM node:18-alpine

WORKDIR /app

# Copy only production dependencies
COPY package*.json ./
RUN npm install --production

# Copy compiled code from builder
COPY --from=builder /app/dist ./dist

# Copy any other runtime files (like .env if needed)
# COPY --from=builder /app/.env .  # optional

# Expose the NestJS default port
EXPOSE 4400

# Run your app
CMD ["node", "dist/main.js"]
