# Stage 1: Builder
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package.json and package-lock.json from src/
COPY src/package*.json ./src/

# Set working directory to /app/src
WORKDIR /app/src

# Install all dependencies (dev + prod)
RUN npm install

# Copy all source code into /app/src
COPY src/ ./

# Build the app
RUN npm run build


# Stage 2: Production
FROM node:18-alpine

WORKDIR /app

# Copy package.json and lock file again from builder stage
COPY src/package*.json ./src/

# Set working directory to /app/src again
WORKDIR /app/src

# Install production dependencies only
RUN npm install --omit=dev

# Copy the built output from builder
COPY --from=builder /app/src/dist ./dist

# Expose the port your app uses
EXPOSE 4005

# Start the app
CMD ["node", "dist/main.js"]
