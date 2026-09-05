# Value-Aligned Bluesky Feed Generator

An applied AI project that analyses Bluesky posts through Schwartz's theory of 19 basic human values and generates personalised feeds from user-defined preferences.

The project extends the value-ranking system developed for my [Computer Science thesis](https://github.com/davide-ferrara/feed-analysis), which replicated a CHI '26 study on Instagram.

## What it does

```text
Bluesky posts -> collection -> multimodal AI classification -> value scores -> personalised feed
```

- Collects posts and metadata through the Bluesky AT Protocol.
- Classifies text and images across 19 Schwartz values on a 0 to 6 scale.
- Stores posts, classifications and model reasoning in SQLite.
- Ranks content with a weighted score based on the user's value preferences.
- Serves the resulting feed through an HTTP service and web interface.

## Screenshots

### Conservative preferences

| Preference sliders | Resulting feed |
|---|---|
| ![Conservative preference sliders](docs/conservative_sliders.png) | ![Conservative personalised feed](docs/conservative_feed.png) |

### Progressive preferences

| Preference sliders | Resulting feed |
|---|---|
| ![Progressive preference sliders](docs/progressist_sliders.png) | ![Progressive personalised feed](docs/progressist_feed.png) |

## Value framework

| Higher-order cluster | Example values |
|---|---|
| Openness to change | Self-direction, stimulation, hedonism |
| Self-enhancement | Achievement, power, face, resources |
| Conservation | Security, conformity, tradition |
| Self-transcendence | Benevolence, universalism, care for nature |

See the complete [19-value framework](docs/values_table.png).

## Project structure

```text
feed-generator/   CLI for collecting and classifying posts
feed-service/     HTTP feed generator and web application
db/               shared SQLite data layer
pkg/schwartz/     shared value types and scoring structures
data-analysis/    Python analysis and visualisation scripts
```

The main application is written in Go. The classifier can use configured hosted AI models through OpenRouter or SiliconFlow, while SQLite provides local persistence.

## Running the classifier

Requires Go 1.26 and a Bluesky app password.

Create `feed-generator/.env`:

```env
BSKY_HANDLE=your-handle.bsky.social
BSKY_APP_PASSWORD=your-app-password
OPEN_ROUTER_KEY=your-openrouter-key
SILICONFLOW_API_KEY=your-siliconflow-key
```

Build the CLI:

```bash
cd feed-generator
make build
```

Collect posts from the configured profiles:

```bash
./bin/feedgen -total 100 collect-profiles
```

Classify the collected posts:

```bash
./bin/feedgen -model=gpt-4o-mini analyze
```

Flags must appear before the subcommand because the CLI uses Go's standard `flag` package.

## Running the feed service

Configure the service as described in [`feed-service/README.md`](feed-service/README.md), then run:

```bash
cd feed-service
make run
```

## License

[MIT](LICENSE)
