# Changelog

All notable changes to awl-bootstrap-node are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Before 1.0,
a minor version may contain breaking changes. They are listed under **Changed** or
**Removed**.

## [0.11.0] - 2026-10-26

### Added

- **Prometheus metrics** at `/metrics` on the HTTP API address, `127.0.0.1:9010` by
  default. They cover the relay service, AutoNAT, connections and the resource manager
  from libp2p, the DHT requests the node serves (`libp2p_kad_dht_*`), the Go runtime and
  the node itself (`awl_bootstrap_*`): connected peers and DHT requests by the app and
  version of the client, how long relayed connections last, the DHT routing table,
  traffic by protocol and the peerstore size. The README lists the metrics and has a
  Prometheus job to collect them. The output also has `awl_*` metrics with zero values
  from the awl client code: drop them when collecting, as the job in the README does.
  ([#17](https://github.com/anywherelan/awl-bootstrap-node/pull/17),
  [#18](https://github.com/anywherelan/awl-bootstrap-node/pull/18))
- `DisablePprof: true` in the config turns off the pprof endpoints under
  `/api/v0/debug/pprof/`. They stay on by default, so keep `HttpListenAddress` on
  `127.0.0.1`. ([#17](https://github.com/anywherelan/awl-bootstrap-node/pull/17))
- `/api/v0/debug/p2p_info` shows the peer ID of the node.
  ([#16](https://github.com/anywherelan/awl-bootstrap-node/pull/16))

### Changed

- **When upgrading, delete the `peerstore` directory.** Stop the node, replace the
  binary, delete `peerstore` next to `config.yaml` and start the node. The peer ID stays
  the same: the key is in `config.yaml`. The node fills the directory again within a few
  minutes. If you keep the old directory, the first start converts it and needs a lot of
  memory: a node with 1 GB of RAM and a 1.1 GB peerstore ran out of memory once and was
  restarted by systemd. ([#18](https://github.com/anywherelan/awl-bootstrap-node/pull/18))
- The HTTP API compresses responses with gzip when the client asks for it, as
  Prometheus does. `/metrics` takes about 14 KB instead of 164 KB.
  ([#18](https://github.com/anywherelan/awl-bootstrap-node/pull/18))
- The HTTP API has timeouts. A client has 10 seconds to send the request headers and
  20 seconds to send the whole request. An idle connection is closed after 2 minutes.
  Sending a response has no limit, so 30-second pprof profiles still work.
  ([#18](https://github.com/anywherelan/awl-bootstrap-node/pull/18))
- Without `LoggerLevel` in the config, the node logs at `info` instead of `debug`.
  ([#15](https://github.com/anywherelan/awl-bootstrap-node/pull/15))
- Updated awl to v0.20.0, go-libp2p to v0.50.0, go-libp2p-kad-dht to v0.42.2, Go to 1.26
  and other dependencies.
  ([#15](https://github.com/anywherelan/awl-bootstrap-node/pull/15),
  [#16](https://github.com/anywherelan/awl-bootstrap-node/pull/16),
  [#18](https://github.com/anywherelan/awl-bootstrap-node/pull/18))

### Fixed

- **The peerstore no longer grows to gigabytes.** Old addresses of peers stayed on disk
  for days, and the `peerstore` directory reached 1.1 GB on a node with about 350 peers.
  After the upgrade it took 14 MB on the same node. The node also uses about 3 times less
  memory: around 100 MB instead of 300 MB.
  ([#18](https://github.com/anywherelan/awl-bootstrap-node/pull/18))
- The node didn't close the peerstore on stop, and 3 seconds to stop were not enough:
  it exited with `exit timeout reached`. Now it closes the peerstore and has up to 15
  seconds to stop. ([#15](https://github.com/anywherelan/awl-bootstrap-node/pull/15),
  [#18](https://github.com/anywherelan/awl-bootstrap-node/pull/18))

[0.11.0]: https://github.com/anywherelan/awl-bootstrap-node/compare/v0.10.0...v0.11.0
