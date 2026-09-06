# kimi-quota-monitor

Polls the [Kimi Code](https://www.kimi.com/code/) quota API at a configurable interval and publishes quota information to an MQTT broker. Built to feed the [esp32-c6-llm-quota](https://github.com/jsvensson/esp32-c6-llm-quota) display.

## How it works

1. Every `POLL_INTERVAL` (defined as [Go duration strings](https://pkg.go.dev/time#ParseDuration)), the service calls `GET https://api.kimi.com/coding/v1/usages` with the API key.
2. It maps the response to a JSON object that reports the remaining quota percentage and reset timestamp for each window, and publishes it as a **retained** QoS 1 message to `MQTT_TOPIC`:

   ```json
   {"5h": {"pct": 0, "resets_at": 1756340000}, "7d": {"pct": 7, "resets_at": 1756340000}}
   ```

   - `7d`: weekly remaining percentage and reset time (`usage` in the API response)
   - `5h`: 5-hour rolling-window remaining percentage and reset time (`limits` in the API response); omitted if the API reports no such window
3. By default, the payload is only published when it differs from the last published one — the retained message keeps consumers updated. Set `MQTT_REPUBLISH=true` to publish every interval regardless. Fetch or publish failures are logged and retried on the next interval.
4. If `HTTP_ADDR` is set, the same JSON body is also served at `GET http://<HTTP_ADDR>/quota`. The endpoint returns `503 Service Unavailable` until the first successful poll.

> Note: the payload format has changed from `[used, limit]` arrays to a single percentage, and now to `{"pct": <remaining>, "resets_at": <unix epoch>}` objects. The ESP32 consumer must be updated to expect objects.

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
| `MQTT_REPUBLISH` | `false` | Publish every interval even when the payload is unchanged; by default only changes are published since the message is retained |
| `HTTP_ADDR` | — | Optional listen address for the HTTP endpoint, e.g. `:8080`; unset disables it |
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

To also expose the HTTP endpoint, add `-p 8080:8080 -e HTTP_ADDR=:8080`.

The Dockerfile sets example defaults for the non-secret variables (`POLL_INTERVAL`, `MQTT_TOPIC`, and so on). Override them with `-e` or `--env-file`.

## Verify

Watch the broker to confirm messages arrive:

```sh
mosquitto_sub -h 192.168.1.10 -t quota/llm -v
```

If `HTTP_ADDR` is set, fetch the same payload over HTTP:

```sh
curl localhost:8080/quota
```

## Test

```sh
just test
```
