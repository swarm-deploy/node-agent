# Node Agent

Node Agent for Docker Swarm nodes.

## Features

- Clean old containers and images
- Collect volume usage metrics

Docker API connection settings are read from standard Docker environment variables
(e.g. `DOCKER_HOST`, `DOCKER_TLS_VERIFY`, `DOCKER_CERT_PATH`).

## Run

```yaml
services:
  node-agent:
    image: swarmdeployorg/node-agent:0.1.0
    environment:
      DOCKER_HOST: "unix:///var/run/docker.sock"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    restart: unless-stopped
    deploy:
      mode: global
```
