# locabft — CometBFT fork by [Locality Network](https://github.com/localitynetwork)

> **This is a fork of [CometBFT v0.38.17](https://github.com/cometbft/cometbft/tree/v0.38.17).**
> It is used as the consensus engine for the [loca](https://github.com/localitynetwork/loca)
> blockchain. The module path is kept as `github.com/cometbft/cometbft` — import via a
> `replace` directive in your `go.mod`:
>
> ```go
> github.com/cometbft/cometbft v0.38.17 => github.com/localitynetwork/locabft v0.38.17-loca.0
> ```

## Changes vs upstream CometBFT v0.38.17

### Problem: infinite empty-block loop on Cosmos SDK chains

The upstream `needProofBlock()` function returns `true` whenever the previous block's
AppHash differs from the header. On Cosmos SDK chains, the `BeginBlocker` modules (mint,
distribution, slashing, staking) modify the AppHash on **every** block — including empty
ones. With `create_empty_blocks = false`, this caused CometBFT to produce empty blocks in
an infinite loop, ignoring the configuration entirely.

### Fix: `needCommitBlock` in `consensus/state.go`

Replaces `needProofBlock` with a simpler function that only forces a block in three cases:

1. **Genesis height** — so the initial AppHash is recorded in a block header.
2. **After restart** — one empty "warm-up" block forces `FinalizeBlock`, refreshing Cosmos
   SDK module state (mint, staking, distribution) before user transactions arrive. Without
   this, txs submitted right after a node restart may be rejected against stale state.
3. **After a block with transactions** — one empty "commit" block records the resulting
   AppHash in the next header. Because this block has `NumTxs==0`, the rule does not apply
   to the block after it, breaking the loop automatically.

After case 3, the node waits for new transactions (or the keepalive interval expires).

### Keepalive (`create_empty_blocks_interval`)

When `create_empty_blocks = false` and `create_empty_blocks_interval` is set (e.g. `"1h"`),
a timeout is scheduled each time the node enters wait mode. When it fires, one empty block
is produced to keep the chain alive and prevent validator downtime penalties. The node then
returns to wait mode and schedules the next keepalive.

### Modified files

| File | Change |
|------|--------|
| `consensus/state.go` | Replace `needProofBlock` with `needCommitBlock`; add `startupBlockCommitted` warm-up field; add `[locabft]` log messages |
| `state/execution.go` | Add `startupBlockCommitted` field to `BlockExecutor` (inactive — kept for reference) |

All fork-specific code is tagged with `locabft:` in comments and `[locabft]` in log
messages, making it easy to identify changes with `grep` and to filter logs on validators:

```bash
docker logs val-hana 2>&1 | grep "\[locabft\]"
```

| Log message | Meaning |
|-------------|---------|
| `[locabft] startup block: warming ABCI app state` | First block after restart (warm-up) |
| `[locabft] startup block: ABCI app state ready` | ABCI app state refreshed |
| `[locabft] waiting for txs` | Node idle, waiting for transactions |
| `[locabft] txs available, proposing block` | New txs in mempool |
| `[locabft] commit block: recording AppHash after txs` | Empty block to record AppHash |
| `[locabft] keepalive: interval elapsed, proposing empty block` | Periodic keepalive block |

---

# CometBFT

[Byzantine-Fault Tolerant][bft] [State Machine Replication][smr]. Or
[Blockchain], for short.

[![Version][version-badge]][version-url]
[![API Reference][api-badge]][api-url]
[![Go version][go-badge]][go-url]
[![Discord chat][discord-badge]][discord-url]
[![License][license-badge]][license-url]
[![Sourcegraph][sg-badge]][sg-url]

| Branch  | Tests                                          | Linting                                     |
|---------|------------------------------------------------|---------------------------------------------|
| main    | [![Tests][tests-badge]][tests-url]             | [![Lint][lint-badge]][lint-url]             |
| v0.38.x | [![Tests][tests-badge-v038x]][tests-url-v038x] | [![Lint][lint-badge-v038x]][lint-url-v038x] |
| v0.37.x | [![Tests][tests-badge-v037x]][tests-url-v037x] | [![Lint][lint-badge-v037x]][lint-url-v037x] |
| v0.34.x | [![Tests][tests-badge-v034x]][tests-url-v034x] | [![Lint][lint-badge-v034x]][lint-url-v034x] |

CometBFT is a Byzantine Fault Tolerant (BFT) middleware that takes a
state transition machine - written in any programming language - and securely
replicates it on many machines.

It is a fork of [Tendermint Core][tm-core] and implements the Tendermint
consensus algorithm.

For protocol details, refer to the [CometBFT Specification](./spec/README.md).

For detailed analysis of the consensus protocol, including safety and liveness
proofs, read our paper, "[The latest gossip on BFT
consensus](https://arxiv.org/abs/1807.04938)".

## Documentation

Complete documentation can be found on the
[website](https://docs.cometbft.com/).

## Releases

Please do not depend on `main` as your production branch. Use
[releases](https://github.com/cometbft/cometbft/releases) instead.

If you intend to run CometBFT in production, we're happy to help. To contact
us, in order of preference:

- [Create a new discussion on
  GitHub](https://github.com/cometbft/cometbft/discussions)
- Reach out to us via [Telegram](https://t.me/CometBFT)
- [Join the Cosmos Network Discord](https://discord.gg/interchain) and
  discuss in
  [`#cometbft`](https://discord.com/channels/669268347736686612/1069933855307472906)

More on how releases are conducted can be found [here](./RELEASES.md).

## Security

To report a security vulnerability, see our [bug bounty
program](https://hackerone.com/cosmos). For examples of the kinds of bugs we're
looking for, see [our security policy](SECURITY.md).

## Minimum requirements

| CometBFT version | Requirement | Notes             |
|------------------|-------------|-------------------|
| main             | Go version  | Go 1.22 or higher |
| v0.38.x          | Go version  | Go 1.22 or higher |
| v0.37.x          | Go version  | Go 1.22 or higher |
| v0.34.x          | Go version  | Go 1.12 or higher |

### Install

See the [install guide](./docs/guides/install.md).

### Quick Start

- [Single node](./docs/guides/quick-start.md)
- [Local cluster using docker-compose](./docs/networks/docker-compose.md)

## Contributing

Please abide by the [Code of Conduct](CODE_OF_CONDUCT.md) in all interactions.

Before contributing to the project, please take a look at the [contributing
guidelines](CONTRIBUTING.md) and the [style guide](STYLE_GUIDE.md). You may also
find it helpful to read the [specifications](./spec/README.md), and familiarize
yourself with our [Architectural Decision Records
(ADRs)](./docs/architecture/README.md) and [Request For Comments
(RFCs)](./docs/rfc/README.md).

## Versioning

### Semantic Versioning

CometBFT uses [Semantic Versioning](http://semver.org/) to determine when and
how the version changes. According to SemVer, anything in the public API can
change at any time before version 1.0.0

To provide some stability to users of 0.X.X versions of CometBFT, the MINOR
version is used to signal breaking changes across CometBFT's API. This API
includes all publicly exposed types, functions, and methods in non-internal Go
packages as well as the types and methods accessible via the CometBFT RPC
interface.

Breaking changes to these public APIs will be documented in the CHANGELOG.

### Upgrades

In an effort to avoid accumulating technical debt prior to 1.0.0, we do not
guarantee that breaking changes (i.e. bumps in the MINOR version) will work with
existing CometBFT blockchains. In these cases you will have to start a new
blockchain, or write something custom to get the old data into the new chain.
However, any bump in the PATCH version should be compatible with existing
blockchain histories.

For more information on upgrading, see [UPGRADING.md](./UPGRADING.md).

### Supported Versions

Because we are a small core team, we have limited capacity to ship patch
updates, including security updates. Consequently, we strongly recommend keeping
CometBFT up-to-date. Upgrading instructions can be found in
[UPGRADING.md](./UPGRADING.md).

Currently supported versions include:

- v0.38.x: CometBFT v0.38 introduces ABCI 2.0, which implements the entirety of
  ABCI++
- v0.37.x: CometBFT v0.37 introduces ABCI 1.0, which is the first major step
  towards the full ABCI++ implementation in ABCI 2.0
- v0.34.x: The CometBFT v0.34 series is compatible with the Tendermint Core
  v0.34 series

## Resources

### Libraries

- [Cosmos SDK](http://github.com/cosmos/cosmos-sdk); A framework for building
  applications in Golang
- [Tendermint in Rust](https://github.com/informalsystems/tendermint-rs)
- [ABCI Tower](https://github.com/penumbra-zone/tower-abci)

### Applications

- [Cosmos Hub](https://hub.cosmos.network/)
- [Terra](https://www.terra.money/)
- [Celestia](https://celestia.org/)
- [Anoma](https://anoma.network/)
- [Vocdoni](https://docs.vocdoni.io/)

### Research

Below are links to the original Tendermint consensus algorithm and relevant
whitepapers which CometBFT will continue to build on.

- [The latest gossip on BFT consensus](https://arxiv.org/abs/1807.04938)
- [Master's Thesis on Tendermint](https://atrium.lib.uoguelph.ca/xmlui/handle/10214/9769)
- [Original Whitepaper: "Tendermint: Consensus Without Mining"](https://tendermint.com/static/docs/tendermint.pdf)

## Join us

CometBFT is currently maintained by [Informal
Systems](https://informal.systems). If you'd like to work full-time on CometBFT,
[we're hiring](https://informal.systems/careers)!

Funding for CometBFT development comes primarily from the [Interchain
Foundation](https://interchain.io), a Swiss non-profit. Informal Systems also
maintains [cometbft.com](https://cometbft.com).

[bft]: https://en.wikipedia.org/wiki/Byzantine_fault_tolerance
[smr]: https://en.wikipedia.org/wiki/State_machine_replication
[Blockchain]: https://en.wikipedia.org/wiki/Blockchain
[version-badge]: https://img.shields.io/github/v/release/cometbft/cometbft.svg
[version-url]: https://github.com/cometbft/cometbft/releases/latest
[api-badge]: https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
[api-url]: https://pkg.go.dev/github.com/cometbft/cometbft
[go-badge]: https://img.shields.io/badge/go-1.22-blue.svg
[go-url]: https://github.com/moovweb/gvm
[discord-badge]: https://img.shields.io/discord/669268347736686612.svg
[discord-url]: https://discord.gg/interchain
[license-badge]: https://img.shields.io/github/license/cometbft/cometbft.svg
[license-url]: https://github.com/cometbft/cometbft/blob/main/LICENSE
[sg-badge]: https://sourcegraph.com/github.com/cometbft/cometbft/-/badge.svg
[sg-url]: https://sourcegraph.com/github.com/cometbft/cometbft?badge
[tests-url]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml
[tests-url-v038x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.38.x
[tests-url-v037x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.37.x
[tests-url-v034x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.34.x
[tests-badge]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=main
[tests-badge-v038x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.38.x
[tests-badge-v037x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.37.x
[tests-badge-v034x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.34.x
[lint-badge]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=main
[lint-badge-v034x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.34.x
[lint-badge-v037x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.37.x
[lint-badge-v038x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.38.x
[lint-url]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml
[lint-url-v034x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml?query=branch%3Av0.34.x
[lint-url-v037x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml?query=branch%3Av0.37.x
[lint-url-v038x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml?query=branch%3Av0.38.x
[tm-core]: https://github.com/tendermint/tendermint
