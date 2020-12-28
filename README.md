# Websocket Proxy

[![Docker Image Version](https://img.shields.io/docker/v/lucascorpion/websocket-proxy?sort=semver)](https://hub.docker.com/r/lucascorpion/websocket-proxy)
[![Docker Image Size](https://img.shields.io/docker/image-size/lucascorpion/websocket-proxy?sort=semver)](https://hub.docker.com/r/lucascorpion/websocket-proxy)
[![Docker Pulls](https://img.shields.io/docker/pulls/lucascorpion/websocket-proxy)](https://hub.docker.com/r/lucascorpion/websocket-proxy)

A websocket proxy with local discovery for piping data between two clients.

The proxy differentiates between two kinds of clients: agents and remotes. Agents are clients that listen for messages, and optionally reply to them. Remotes are clients that want to connect to an agent.

## API

### `/agents`

A discovery endpoint which lists all agents in the local network (based on the request's public IP address).

### `/ws/agent` and `/ws/remote`

The websocket endpoints for the agent and remote clients, respectively.

## Websocket Protocol

### Agent

After connecting, the agent should send a message with a JSON object. This object can contain a `meta` field with freeform metadata about the agent. For example:

```json
{
  "meta": {
    "version": "1.0.0",
    "platform": "linux"
  }
}
```

The proxy will then assign a unique id to this agent, which remotes can use to identify it. It will send the full agent object back to the agent:

```json
{
  "id": "4cab7bec-dfc2-48a9-a8c9-406118b4242f",
  "meta": {
    "version": "1.0.0",
    "platform": "linux"
  }
}
```

### Remote

After connecting, the agent should send a message with a JSON object. This object should contain an `id` field with the id of the agent to connect to. For example:

```json
{
  "id": "4cab7bec-dfc2-48a9-a8c9-406118b4242f"
}
```

### Diagram

![Sequence diagram](docs/sequence.jpg)
