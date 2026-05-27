# Remna Traffic Drifter Bot

<img width="300" alt="drifter-preview" src="https://github.com/user-attachments/assets/0aa1956b-e23f-4e41-aa5c-f64c45fe9194" />

[![DockerHub](https://img.shields.io/badge/DockerHub-remna--traffic--drifter--bot-blue?style=for-the-badge)](https://hub.docker.com/r/sedyh/remna-traffic-drifter-bot)

Background monitor for Remnawave: polls the panel API and sends Telegram alerts when paid users have the wrong traffic reset strategy or have not had a calendar reset for too long.

Expected strategy and stale-reset threshold are configured per tariff tag in `TAG_RULES`.

## What it checks

Polls `GET /api/users` (paginated).

| Type | Condition |
|------|-----------|
| `wrong_strategy` | `trafficLimitStrategy` does not match the strategy from `TAG_RULES` for the user tag |
| `stale_reset` | expected strategy is `MONTH`, traffic was used, but `lastTrafficResetAt` is empty or older than the stale threshold |

Filters:

- only tags listed in `TAG_RULES` are checked
- `WATCH_UNTAGGED_LIMITED=true` - also checks limited users without a tag (`UNTAGGED_LIMITED_RULE`)
- unlimited users (`trafficLimitBytes=0`) are skipped
- `EXPIRED` users are skipped
- tags not listed in `TAG_RULES` (for example `TRIAL`) are skipped
- `stale_reset` with empty `lastTrafficResetAt` is not alerted if the account is younger than the stale threshold

Alerts link the subscription name to the user page in the panel. You must be logged into the panel (session cookie after sign-in).

## Docker (minimal setup)

Edit `environment` in [drifter.compose.yaml](drifter.compose.yaml) and run:

```bash
docker compose -f drifter.compose.yaml up -d
```

Image on [Docker Hub](https://hub.docker.com/r/sedyh/remna-traffic-drifter-bot): `sedyh/remna-traffic-drifter-bot`.

Persistent data is stored in the `/data` volume (`STATE_PATH`, `TELEGRAM_OFFSET_PATH` by default).

### Required variables

| Variable | Description |
|----------|-------------|
| `PANEL_URL` | Remnawave panel base URL |
| `PANEL_TOKEN` | API bearer token |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token |
| `TELEGRAM_CHAT_IDS` | Comma-separated chat IDs; use `chat_id:topic_id` for forum topics |
| `TAG_RULES` | Per-tag rules: `TAG:STRATEGY` or `TAG:STRATEGY:STALE` (comma-separated) |

`TAG_RULES` examples:

- `TAG_TARIFF_MIN:MONTH:35d` - expect `MONTH`, stale reset after `35d`
- `TAG_TARRIF_PRO:MONTH` - expect `MONTH`, stale threshold from `STALE_RESET_AFTER`
- `MAX:NO_RESET` - expect `NO_RESET`, no stale-reset checks

Strategies: `MONTH`, `NO_RESET`. Stale accepts Go durations (`35d`, `720h`).

### Optional variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_SEND_INTERVAL` | `3s` | Delay between outbound messages |
| `TELEGRAM_PROXY_URL` | - | HTTP(S) proxy for Telegram API |
| `POLL_INTERVAL` | `15m` | Panel poll interval |
| `STALE_RESET_AFTER` | `35d` | Default stale threshold when not set per tag |
| `UNTAGGED_LIMITED_RULE` | `MONTH` | Rule for limited users without a tag (`STRATEGY` or `STRATEGY:STALE`) |
| `WATCH_UNTAGGED_LIMITED` | `true` | Watch limited users without a tag |
| `PAGE_SIZE` | `500` | `/api/users` page size |
| `STATE_PATH` | `/data/state.json` | Alert deduplication state file |
| `TELEGRAM_OFFSET_PATH` | `/data/telegram_offset` | Long polling offset file |

### `docker run` ([Docker Hub](https://hub.docker.com/r/sedyh/remna-traffic-drifter-bot))

```bash
docker run -d --name traffic-drifter --restart unless-stopped \
  -e PANEL_URL=https://panel.example.com \
  -e PANEL_TOKEN=secret \
  -e TELEGRAM_BOT_TOKEN=secret \
  -e TELEGRAM_CHAT_IDS=-1001234567890:289 \
  -e TAG_RULES=TAG_TARIFF_MIN:MONTH:35d,TAG_TARRIF_PRO:MONTH:35d \
  -v traffic-drifter-data:/data \
  sedyh/remna-traffic-drifter-bot:latest
```

## Local run

```bash
cp .env.example .env
go run ./cmd/traffic_drifter
```

## Development

```bash
task test
task build
task image
```

Docker Hub publish: push a `v*` git tag (`secrets.DOCKERHUB_USERNAME`, `secrets.DOCKERHUB_TOKEN`).
