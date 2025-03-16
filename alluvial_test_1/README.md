# Problem description

Build a Golang service that proxies the eth_getBalance RPC endpoint from the Ethereum execution layer.

The key attribute we are looking for is the development of a High Availability service (HA).

Candidates must demonstrate:
- A highly available service.
- The ability to have multiple clients sat behind the proxy.
- A strategy for handling inconsistent data returned by the set of clients.
- An approach for validating their design decisions.

Note that: You will need to sign up for multiple execution node clients, there are many freemium API gateways, such as Infura, Alchemy, Chainstack, Tenderly, etc.

Example:

```
$ GET localhost:8080/eth/balance/0xfe3b557e8fb62b89f4916b721be55ceb828dbd73
{
"balance": "1000"
}
```

## Bonus Criteria

The service should expose Prometheus metrics at /metrics.
The service should expose a liveness and readiness HTTP endpoint.

## Extra information

The service should be architected and coded using idiomatic Go.
You are free to use extra dependencies from the Golang ecosystem.

### Recommended resources

eth_gasBalance method: https://docs.infura.io/infura/networks/ethereum/json-rpc-methods/eth_getbalance

Kubernetes probes https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/

Prometheus instrumentation in Go
https://prometheus.io/docs/guides/go-application/


# How to build and run locally
Running basic setup with docker compose: `docker compose up -d`.
After this service is available locally on port 8080. 
Verify with curl: `curl --location 'http://localhost:8080/health/live'`
