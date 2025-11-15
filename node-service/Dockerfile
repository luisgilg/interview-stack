# syntax=docker/dockerfile:1
FROM node:20-alpine
WORKDIR /usr/src/app

COPY package.json ./
RUN npm install --production && npm cache clean --force

COPY cmd ./cmd
COPY internal ./internal
COPY config.yaml ./config.yaml
COPY wait-for.sh ./wait-for.sh

RUN apk add --no-cache netcat-openbsd

ENV NODE_SERVICE_PORT=8082
EXPOSE 8082

CMD ["/bin/sh", "-c", "./wait-for.sh ${POSTGRES_HOST:-postgres} ${POSTGRES_PORT:-5432} ./wait-for.sh ${MONGO_HOST:-mongo} ${MONGO_PORT:-27017} ./wait-for.sh ${REDIS_HOST:-redis} ${REDIS_PORT:-6379} npm start"]
