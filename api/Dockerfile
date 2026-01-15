# Use an official Node.js image v25.3
FROM node@sha256:a2f09f3ab9217c692a4e192ea272866ae43b59fabda1209101502bf40e0b9768

WORKDIR /src

# Copy package.json and package-lock.json first for caching
COPY src/package*.json ./

# Install dependencies
RUN npm install --production

# Copy the rest of the app source code
COPY src/ .
COPY migrations/ /src/migrations/

# Expose the port
EXPOSE 8080

CMD ["node", "index.js"]
