# awl-bootstrap-node

This project implements a bootstrap node — a well-known, stable peer that helps new nodes join the P2P network by providing initial connection information. It supports DHT, traffic relaying, and NAT traversal. [Here's](https://github.com/anywherelan/awl#how-it-works) a bit more information about it.

The Anywherelan project maintains a list of initial bootstrap nodes, which can be found [here](https://github.com/anywherelan/awl/blob/v0.12.0/config/other.go#L32-L40). However, you can help the community by setting up your own!

## Getting started

To build this project, you will need a Go compiler with a version of 1.26 or higher.

```bash
CGO_ENABLED=0 go build -trimpath
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

The node exposes Prometheus metrics at `/metrics` on its HTTP API address,
`HttpListenAddress` in the config, `127.0.0.1:9010` by default:

```bash
curl http://127.0.0.1:9010/metrics
```

They include:

- libp2p metrics: `libp2p_relaysvc_*`, `libp2p_autonatv2_*`, `libp2p_swarm_*`,
  `libp2p_rcmgr_*`, `libp2p_identify_*`, `libp2p_holepunch_*` and others;
- DHT requests served by the node: `libp2p_kad_dht_*`;
- Go runtime and process metrics: `go_*`, `process_*`;
- metrics of the node itself, `awl_bootstrap_*`:

| Metric | What it shows |
|---|---|
| `awl_bootstrap_node_info`, `awl_bootstrap_node_uptime_seconds` | version, peer ID and uptime |
| `awl_bootstrap_p2p_peers_connected` | connected peers by `app`, `version`, `os` and `arch` from their user agent |
| `awl_bootstrap_p2p_dht_requests_total` | DHT requests by `message_type` and by `app` and `version` of the sender |
| `awl_bootstrap_p2p_relay_circuit_duration_seconds` | how long relayed connections last, failed attempts included |
| `awl_bootstrap_p2p_dht_routing_table_size` | DHT routing table size |
| `awl_bootstrap_p2p_bootstrap_peers_connected` | connected bootstrap peers from the config |
| `awl_bootstrap_p2p_bandwidth_bytes_total` | traffic by `protocol` and `direction`: data of libp2p streams, without TCP, QUIC and IP overhead |
| `awl_bootstrap_peerstore_peers_with_addrs`, `awl_bootstrap_peerstore_peers_with_keys` | peerstore size |

The output also has `awl_*` metrics with zero values. They come from the awl client
code that the node reuses. Drop them when scraping, see the example below.

### Grafana and Prometheus

The [`awl`](https://github.com/anywherelan/awl) repository has a Prometheus and
Grafana stack in [`monitoring/`](https://github.com/anywherelan/awl/tree/master/monitoring).
It runs with docker compose and has the official libp2p dashboards. The node has
the same `libp2p_*` metrics, so the stack works here as is. For a bootstrap node
the most useful dashboards are **relaysvc**, **autonatv2**, **swarm** and
**resource-manager**.

To collect metrics from this node, add a job to `monitoring/prometheus.yml` with
its HTTP address. The node uses port `9010` by default, awl uses `8639`:

```yaml
  - job_name: awl-bootstrap
    metrics_path: /metrics
    static_configs:
      - targets:
          - host.docker.internal:9010
    # drop the zero awl_* metrics of the awl client code
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: awl_(node|p2p|peers|vpn|dns|socks5)_.+
        action: drop
```

## Profiling

pprof endpoints are served under `/api/v0/debug/pprof/` and are enabled by default.
Set `DisablePprof: true` in the config to turn them off.