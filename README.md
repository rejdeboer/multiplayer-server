# Multiplayer server (WIP)

This project is about creating a centralized cloud-based collaborative text editor using CRDTs. For a great explanation about CRDTs, I recommend watching [Martin Kleppmann's video](https://www.youtube.com/watch?v=x7drE24geUw) about the subject.

The project consists of 2 separate servers:

1. An HTTP server written in Golang. The code can be found in the `internal` and `pkg` directories
2. A WebSocket server written in Rust using Axum. The code can be found in the `websocket` directory

The HTTP server currently supports the following features:

- Logging with zerolog
- File storage in Azure Blob Storage
- Simple JWT authentication, users stored in Postgres DB
- Document creation and listing endpoint

The WebSocket server currently supports the following features:

- [y-crdt](https://github.com/y-crdt/y-crdt) document persistence in Postgres DB
- WebSocket endpoint to receive real-time updates

The client code used to connect to this server can be found at [multiplayer-client](https://github.com/rejdeboer/multiplayer-client)

## Deployment

For an example of how to deploy this app to Azure on a Kubernetes cluster, have a look at [this repo](https://github.com/rejdeboer/multiplayer-deployment)
