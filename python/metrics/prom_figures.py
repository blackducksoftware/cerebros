import prometheus

from datetime import datetime, timedelta, timezone
import json
import matplotlib.pyplot as plt
import os
from pathlib import Path
import sys

url, namespace = sys.argv[1:]

prom = prometheus.Prometheus(url)

memory_usage_query = """
sum by(container) (
    container_memory_usage_bytes{{
        job="kubelet",
        namespace="{namespace}",
        container!="POD",
        container!=""
    }}
)
"""

path = "memory_usage.json"
grab_fresh_data = True
if grab_fresh_data:
    start, end = prometheus.interval(hours=4)
    step = 60
    resp = prom.issue_request(
        memory_usage_query.format(namespace=namespace),
        start.timestamp(),
        end.timestamp(),
        step)
    with open(path, "w") as outfile:
        body = resp.json()
        outfile.write(json.dumps(body, indent=2))

    for r in body["data"]["result"]:
        print(r["metric"])

metric = prometheus.Metric.read_file(path)

print(metric.keys, metric.series.keys())

Path("my-pngs").mkdir(exist_ok=True)

for (key, s) in metric.series.items():
    print(key, len(s["values"]), len(s["timestamps"]))
    name = os.path.join("my-pngs", "_".join(key[0]))
    xs = [datetime.utcfromtimestamp(t) for t in s["timestamps"]]
    print(xs)
    ys = [float(x) for x in s["values"]]
    plt.plot(xs, ys)
    print("writing file to {}.png".format(name))
    plt.savefig(name)
    # plt.show()
    plt.clf()
