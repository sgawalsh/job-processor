# Use an official Node.js image
FROM node:20-alpine

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
