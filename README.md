# kimi-quota-monitor

Polls the [Kimi Code](https://www.kimi.com/code/) quota API at a configurable interval and publishes the used/limit pairs to an MQTT broker. Built to feed the [esp32-c6-llm-quota](https://github.com/jsvensson/esp32-c6-llm-quota) display.

## How it works

1. Every `POLL_INTERVAL`, the service calls `GET https://api.kimi.com/coding/v1/usages` with the API key.
2. It maps the response to a JSON object that reports only the current used value for each quota window and publishes it as a **retained** QoS 1 message to `MQTT_TOPIC`:

   ```json
   {"5h": 45, "7d": 93}
   ```

   - `7d`: weekly used value (`usage` in the API response)
   - `5h`: 5-hour rolling-window used value (`limits` in the API response); omitted if the API reports no such window
3. Fetch or publish failures are logged and retried on the next interval. The retained message stays on the broker, so consumers keep the last known values.

> Note: the payload was previously `[used, limit]` arrays. Since Kimi Code reports these windows as simple percentages, it now sends only the used value. The ESP32 consumer must be updated to expect numbers instead of arrays.

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|---|---|---|
| `KIMI_API_KEY` | *(required)* | Kimi Code console API key (`sk-kimi-xxx`) |
| `KIMI_BASE_URL` | `https://api.kimi.com/coding/v1` | API base URL |
| `POLL_INTERVAL` | `5m` | Poll interval (Go duration string) |
| `MQTT_BROKER` | *(required)* | Broker URL, e.g. `tcp://192.168.1.10:1883` |
| `MQTT_CLIENT_ID` | `kimi-quota-monitor` | MQTT client ID |
| `MQTT_TOPIC` | `quota/llm` | Topic to publish to |
| `MQTT_USERNAME` | — | Optional broker username |
| `MQTT_PASSWORD` | — | Optional broker password |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

Get an API key from the Kimi Code console. Note: this is a **Kimi Code** key (`sk-kimi-xxx`), not a Kimi open platform key (`sk-xxx`) — they are not interchangeable.

## Build and run

```sh
just build
KIMI_API_KEY=sk-kimi-xxx MQTT_BROKER=tcp://192.168.1.10:1883 ./kimi-quota-monitor
```

Or in one step: `just run` (reads the same environment variables).

## Docker

```sh
docker build -t kimi-quota-monitor .
docker run --rm \
  -e KIMI_API_KEY=sk-kimi-xxx \
  -e MQTT_BROKER=tcp://192.168.1.10:1883 \
  kimi-quota-monitor
```

The Dockerfile sets example defaults for the non-secret variables (`POLL_INTERVAL`, `MQTT_TOPIC`, and so on). Override them with `-e` or `--env-file`.

## Verify

Watch the broker to confirm messages arrive:

```sh
mosquitto_sub -h 192.168.1.10 -t quota/llm -v
```

## Test

```sh
just test
```
