# Websocket - TCP Proxy

A two-way websocket - TCP proxy server for piping data between two clients.

The proxy differentiates between two kinds of clients: agents and remotes. Agents are TCP clients that listen for messages, and optionally reply to them. Remotes are websocket clients that want to connect to an agent.

## Protocol

Agent (TCP) clients should connect to port `6061`, and remote (websocket) clients should connect to port `6062`, on path `/ws`.

When connecting to the proxy, the first message contains metadata about the agent or remote. For the agent this is a JSON object with the following properties:

| Property   | Description |
|------------|-------------|
| `id`       | A unique id for this agent.
| `version`  | The version of the agent.
| `address`  | The local address (IP and port) at which the agent is running.
| `platform` | The platform the agent is running on, e.g. `windows`, `linux` or `mac`.
| `hostname` | The name of the host the agent is running on.

For the remote this message is only the `id` of the agent to connect to.
