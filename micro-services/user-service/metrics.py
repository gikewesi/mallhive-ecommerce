# metrics.py
import requests
from datetime import datetime

METRICS_ENDPOINT = "http://monitoring.internal.mallhive.com/metrics"

def emit_metric(name: str, value: int = 1, **labels):
    metric_data = {
        "name": name,
        "value": value,
        "labels": labels,
        "timestamp": datetime.utcnow().isoformat()
    }

    try:
        requests.post(METRICS_ENDPOINT, json=metric_data, timeout=2)
    except requests.RequestException:
        # Best-effort, don’t crash main app
        pass
