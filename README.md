# tachyne-gw-java-770

> tachyne is an unofficial fan project, not affiliated with Mojang,
> Microsoft, or Minecraft's developer/publisher in any way. See the
> Disclaimer at the bottom.


Gateway for Java protocols **770–772 (Minecraft 1.21.5–1.21.8)**: terminates
real clients, authorizes every login via **tachyne-access** (fail closed,
30 s verdict cache), attaches the session to a **tachyne-world** pod over the
domain attach protocol, and renders the typed event stream into 770 wire
format via the shared **`tachyne-common/render770`** package. 770→772 are
pure packet-ID remaps, so all three protocols are served from the one
canonical composition.

Clients do not reach this pod directly: **tachyne-ingress** owns the public
port (`<server-ip>:25565`), reads the handshake, and splices matching
protocols here (cluster-internal service `:25570`), prefixing PROXY protocol
v1 so access checks and logs see the real client IP.

## Session pipeline

```
client ⇄ [status | login(access check) → configuration → play]   this repo
              play ⇄ attach session ⇄ tachyne-world :25500 (ATTACH_TOKEN)
```

- **Login/config**: offline-mode Login Success, Set Compression (zlib,
  threshold 256), then the shared config-phase composition from
  `tachyne-common/protocol` — full registries (enchantments included),
  `UpdateTagsPacket`, brand, feature flags. Identical bytes to what the old
  monolith sent a 770 client, guarded by a strict re-parser test in common.
- **Play, world→client** (`internal/gw/session.go`): every attach frame has a
  typed case rendering through `render770` — entities via the per-viewer
  `EntityView` (relative i16 deltas vs absolute resyncs, `NoSync`), chat/boss
  bars/time, survival state, items/windows, sounds/particles/world FX, game
  events/abilities/vehicles, Dimension→Respawn + Teleport→re-center/re-Want.
  Chunks decode from the attach binary body into 770 chunk-data packets
  (block entities included via `ChunkHeader.BEs`).
- **Play, client→world**: movement → `Move` frames + `Want` on center-chunk
  change (view radius 6); dig/place parsed and forwarded as typed frames with
  the prediction sequence **acked locally**; chat/commands forwarded; every
  other gameplay packet lifted via `render770.SID*`/`Parse*` into the typed
  serverbound actions 0x34–0x3f (use-item/entity, window clicks, crafting,
  anvil/enchant, sneak/sprint, respawn, creative slots). Unknown serverbound
  packets are dropped — if a client feature doesn't work, add its parser to
  render770 and its frame to attach; there is no raw fallback.

## Build / run

```bash
go build ./... && go test ./...
go run ./cmd/gw    # env-first config
```

Key env: `TACHYNE_LISTEN`, `TACHYNE_BACKEND` (world attach addr —
`tachyne-world-0.…:25500` in-cluster, `localhost:25500` for local dev
against a local engine), `TACHYNE_ATTACH_TOKEN` (secret
`tachyne-attach-token`), `TACHYNE_ACCESS_URL` + `TACHYNE_ACCESS_TOKEN`
(unset = access checks off, dev only), `POD_NAME` (ordinal → SID). See
`cmd/gw/main.go` for the authoritative list.

## Deploy

CI builds + pushes the gateway image
on every push to main (dind quirk + `REGISTRY_TOKEN` notes in the gw-776
repo's CLAUDE.md); then `kubectl rollout restart` on the StatefulSet. Deploy the world pod first when a
tachyne-common protocol change is involved.

## Known debt

Recipe book not sent (needs a gateway-side builder over a future attach
event); `joinPacket` hardcodes survival at login (Welcome lacks gamemode —
mode-change events correct it after join); sibling gw-776 duplicates the
session code with a translation layer — extraction into common is future
work if a third gateway appears.

## Deployment

`Dockerfile` builds a static Go binary into a minimal image. `deploy/` holds
working Kubernetes manifests (the ones this project actually runs) — treat
them as examples: substitute your own image registry, hostnames, namespaces
and secrets before applying them to your cluster.

## Credits

All protocol rendering comes from the shared `tachyne-common` library — see
its credits (PrismarineJS/minecraft-data, misode/mcmeta, the Minecraft Wiki,
ViaVersion as factual references). This repo itself has no third-party
dependencies beyond that library.

## Development transparency

tachyne is built by its maintainer working with an AI coding agent
(Anthropic's Claude): substantial portions of the implementation were written
by the model under human direction, and every change is reviewed, tested and
deployed by the maintainer. The project's engineering discipline is designed
for exactly this workflow — byte-oracle tests pin the wire format, full test
suites gate every image build, and real-client verification signs off
gameplay. Disclosed here for transparency; judge the code on its behavior.

## License

Licensed under the **Apache License, Version 2.0** — see [LICENSE](LICENSE)
and [NOTICE](NOTICE). Note §6: the license grants no rights to the tachyne
name or any trademarks.

## Disclaimer

tachyne is an unofficial, independent project. It is **not** affiliated with,
endorsed, sponsored, or approved by Mojang Studios, Mojang Synergies AB,
Microsoft Corporation, or any of their subsidiaries — the developer and
publisher of Minecraft have no involvement with this project. "Minecraft" is
a trademark of Mojang Synergies AB. This project contains no Minecraft game
code; all game behavior is independently reimplemented, and data tables are
built from openly licensed community datasets (see Credits).
