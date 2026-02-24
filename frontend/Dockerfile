FROM node:25-alpine3@sha256:b9b5737eabd423ba73b21fe2e82332c0656d571daf1ebf19b0f89d0dd0d3ca93

WORKDIR /app

COPY /src/package*.json ./
RUN npm install

COPY /src .

EXPOSE 5173

CMD ["npm", "run", "dev", "--", "--host"]