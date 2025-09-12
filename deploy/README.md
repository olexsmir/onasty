# Deploy

>[!IMPORTANT] Before deploying:
> 1. Set the environment variables in `docker-compose.yml` and `docker-compose.monitoring.yml` before running them.
> 1. Set your domain in `../infra/caddy/Caddyfile`.

Building the frontend app, so it can be served:
```bash
./build.sh
```

Run the containers:
```bash
docker compose up -d
```

Run the monitoring suite:
```bash
docker compose up -d -f docker-compose.monitoring.yml
```

The monitoring suite is not added to the Caddyfile, so you would need to be in the same network to access it.
