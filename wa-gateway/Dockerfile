FROM node:20-slim
WORKDIR /app

COPY package*.json ./
RUN npm install --omit=dev --no-audit

COPY . .

EXPOSE 5001
ENV PORT=5001
ENV NODE_ENV=production
ENV IS_SERVER=true

CMD ["node", "server.js"]
