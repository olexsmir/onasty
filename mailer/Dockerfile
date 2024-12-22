FROM oven/bun:1 AS builder

WORKDIR /app

COPY package.json bun.lockb tsconfig.json ./
RUN bun install  --frozen-lockfile

COPY src src
RUN bun build ./src/index.ts --target=bun --outfile=dist.js

FROM oven/bun:1
COPY --from=builder /app/dist.js /dist.js
ENTRYPOINT ["bun", "run", "/dist.js"]
