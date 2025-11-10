# This Dockerfile is at apps/user-docs/Dockerfile
# BUT the build context is the root ('.')

FROM node:22-alpine
# Install pnpm
RUN npm install -g pnpm

# Set WORKDIR to the root of the monorepo
WORKDIR /app

# 1. Copy all root manifests and configs
COPY pnpm-lock.yaml .
COPY pnpm-workspace.yaml .
COPY package.json .
COPY turbo.json .

# 2. Copy all package.json files for all apps/packages
COPY apps/frontend/package.json ./apps/frontend/
COPY apps/user-docs/package.json ./apps/user-docs/
COPY packages/eslint-config/package.json ./packages/eslint-config/
COPY packages/typescript-config/package.json ./packages/typescript-config/

# 3. Install ALL dependencies from the root
RUN pnpm install

# 4. Copy all source code
COPY . .

# 5. Set the final WORKDIR to be this app's directory
WORKDIR /app/apps/user-docs

EXPOSE 3000
# Run the start command for this specific app
CMD ["pnpm", "run", "start", "--host", "0.0.0.0"]