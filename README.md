# awl-bootstrap-node

This project implements a bootstrap node — a well-known, stable peer that helps new nodes join the P2P network by providing initial connection information. It supports DHT, traffic relaying, and NAT traversal. [Here's](https://github.com/anywherelan/awl#how-it-works) a bit more information about it.

The Anywherelan project maintains a list of initial bootstrap nodes, which can be found [here](https://github.com/anywherelan/awl/blob/v0.12.0/config/other.go#L32-L40). However, you can help the community by setting up your own!

## Getting started

To build this project, you will need a Go compiler with a version of 1.26 or higher.

```bash
CGO_ENABLED=0 go build
```

First, you need to generate example config.
```bash
./awl-bootstrap-node -generate-config
# you can also provide a config name using arg -config-name=config-2.yaml
```

It will print something like this:
```
Generated example config file: config.yaml
Below are addresses that can be used to connect to this bootstrap node
Replace 127.0.0.1 or ::1 with appropriate IP addresses. If you don't have IPv6 support, just skip it

/ip4/127.0.0.1/tcp/6150/p2p/12D3KooWKubLLHV1WFkfzAcj9h2AuGNGsRkoztQZQVPsL1Khi2Hg
/ip4/127.0.0.1/udp/6150/quic-v1/p2p/12D3KooWKubLLHV1WFkfzAcj9h2AuGNGsRkoztQZQVPsL1Khi2Hg
/ip6/::1/tcp/7250/p2p/12D3KooWKubLLHV1WFkfzAcj9h2AuGNGsRkoztQZQVPsL1Khi2Hg
/ip6/::1/udp/7250/quic-v1/p2p/12D3KooWKubLLHV1WFkfzAcj9h2AuGNGsRkoztQZQVPsL1Khi2Hg
```

This configuration file `config.yaml` is ready to use. It will listen on the 0.0.0.0 address and has default community bootstrap peers.


It needs a `config.yaml` file near the binary to run.
```bash
./awl-bootstrap-node
```

To use this server with your `awl` client, you need to modify `bootstrapPeers` field in the `config_awl.json` file, which is located [here](https://github.com/anywherelan/awl#config-file-location).

Example:
```json
{
  "p2pNode": {
    "peerId": "REDACTED",
    "name": "REDACTED",
    "identity": "REDACTED",
    "bootstrapPeers": [
      "/ip4/127.0.0.1/tcp/6150/p2p/12D3KooWKubLLHV1WFkfzAcj9h2AuGNGsRkoztQZQVPsL1Khi2Hg"
    ]
  },
.... other fields are omitted
}
```

## Metrics

The node exposes Prometheus metrics at `/metrics` on its HTTP API address
(`httpListenAddress`, default `127.0.0.1:9010`):

```bash
curl http://127.0.0.1:9010/metrics
```

This includes libp2p's built-in metrics (`libp2p_relaysvc_*`, `libp2p_autonatv2_*`,
`libp2p_swarm_*`, `libp2p_rcmgr_*`, `libp2p_identify_*`, `libp2p_holepunch_*`, ...),
Go runtime/process metrics (`go_*`, `process_*`), and a few bootstrap-node-specific
gauges (`awl_bootstrap_*`): node info/uptime, DHT routing table size, node bandwidth and
peerstore size. The official libp2p Grafana dashboards can be used as-is.

### Grafana + Prometheus stack

The [`awl`](https://github.com/anywherelan/awl) repository ships a ready-to-use
Prometheus + Grafana monitoring stack under
[`monitoring/`](https://github.com/anywherelan/awl/tree/master/monitoring)
(docker-compose with the official libp2p dashboards). Since the bootstrap node
exposes the same `libp2p_*` metrics, that stack works here without changes — the
most relevant dashboards for a bootstrap node are **relaysvc**, **autonatv2**,
**swarm** and **resource-manager**.

To point it at this node, set the scrape target in `monitoring/prometheus.yml` to
this node's HTTP address (default port `9010` instead of awl's `8639`):

```yaml
  - job_name: awl-bootstrap
    metrics_path: /metrics
    static_configs:
      - targets:
          - host.docker.internal:9010
```

## Profiling

pprof endpoints are served under `/api/v0/debug/pprof/` and are enabled by default.
Set `disablePprof: true` in the config to turn them off.