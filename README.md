# rssgram

**rssgram** is a service for automatically publishing news from RSS feeds to a Telegram channel.

## Description

- Aggregates news from specified RSS feeds.
- Saves new items to a SQLite database.
- Sends new items to a Telegram channel via a bot.
- Collects internal metrics for monitoring.

## How it works

1. Loads configuration from the `config.yaml` file (see example: `config.yaml.example`).
2. Periodically fetches RSS feeds and saves new items to the database.
3. Sends new items to the Telegram channel.
4. Starts an HTTP endpoint `/metrics` on port 2222 for internal metrics.

## Quick Start

### 1. Download a ready-to-use binary
You can download a pre-built binary for your platform from the [Releases](https://github.com/<your-repo>/releases) section on GitHub.

- Download the binary for your OS and architecture
- Place it in your working directory
- Make it executable if needed: `chmod +x rssgram`

### 2. Clone the repository (optional, for config/example)
```sh
git clone <repo_url>
cd rssgram
```

### 3. Create a config
Copy the example:
```sh
cp config.yaml.example config.yaml
```
Edit `config.yaml`, add your RSS feeds and Telegram bot parameters.

### 4. Run
```sh
./rssgram
```

Or use Docker/systemd as described below.

#### Using Docker:
```sh
docker build -t rssgram .
docker run -v $(pwd)/config.yaml:/config.yaml rssgram
```

#### Using systemd (Linux):
- Edit `devops/rssgram.service` for your paths.
- Copy the binary and config to your server.
- Start the service via systemd.

### 5. Metrics

Open [http://localhost:2222/metrics](http://localhost:2222/metrics) to view internal service metrics.

## Tests

```sh
make test
```

---

**Author:** [your name or link] 