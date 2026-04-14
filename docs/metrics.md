# Metrics

Metrics can be used to track how Dockwatch behaves over time.

To use this feature, you have to set an [API token](arguments.md#http_api_token) and [enable the metrics API](arguments.md#http_api_metrics),
as well as creating a port mapping for your container for port `8080`.

The metrics API endpoint is `/v1/metrics`.

## Available Metrics

| Name | Type | Description |
| --- | --- | --- |
| `dockwatch_containers_scanned` | Gauge | Number of containers scanned for changes by dockwatch during the last scan |
| `dockwatch_containers_updated` | Gauge | Number of containers updated by dockwatch during the last scan |
| `dockwatch_containers_failed` | Gauge | Number of containers where update failed during the last scan |
| `dockwatch_scans_total` | Counter | Number of scans since the dockwatch started |
| `dockwatch_scans_skipped` | Counter | Number of skipped scans since dockwatch started |

## Example Prometheus `scrape_config`

```yaml
scrape_configs:
  - job_name: dockwatch
    scrape_interval: 5s
    metrics_path: /v1/metrics
    bearer_token: demotoken
    static_configs:
      - targets:
        - 'dockwatch:8080'
```

Replace `demotoken` with the Bearer token you have set accordingly.

## Demo

The repository contains a demo with prometheus and grafana, available through `docker-compose.yml`. This demo
is preconfigured with a dashboard, which will look something like this:

![grafana metrics](assets/grafana-dashboard.png)
