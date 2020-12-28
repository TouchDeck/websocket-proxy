Go to https://mermaid-js.github.io/mermaid-live-editor to generate and download the image to use in the readme.

```mermaid
sequenceDiagram
    participant Agent
    participant Proxy
    participant Remote

    %% Register the agent.
    Agent->>Proxy: { "meta": { ... } }
    Proxy-->>Agent: { "id": "<id>", "meta": { ... } }

    %% Connect to an agent.
    Remote->>Proxy: { "id": "<id>" }

    Note over Agent,Remote: Connection established, proxy is now fully transparent.

    %% Connected, piping data.
    loop Connected
        Remote->>Agent: Message
        Agent-->>Remote: Response
    end
```
