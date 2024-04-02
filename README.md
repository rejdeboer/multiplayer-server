# Multiplayer server (WIP)

As of writing, this is just a simple golang server with the following features:

- Logging with zerolog
- File storage in Azure Blob Storage
- Simple authentication, users stored in Postgres DB
- WebSocket chat endpoint

This project is about creating a centralized cloud-based collaborative text editor using CRDTs. For a great explanation about CRDTs, I recommend watching [Martin Kleppmann's video](https://www.youtube.com/watch?v=x7drE24geUw) about the subject.

The client code used to connect to this server can be found at [multiplayer-client](https://github.com/rejdeboer/multiplayer-client)

## Deployment

For an example of how to deploy this app to Azure on a Kubernetes cluster, have a look at [this repo](https://github.com/rejdeboer/multiplayer-deployment)
