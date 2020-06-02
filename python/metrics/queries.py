import new_relic
import prometheus

import collections
from datetime import datetime, timedelta, timezone
import math
import matplotlib.pyplot as plt
import os
import sys

# setup:
#   pip3 install pandas
#   pip3 install requests
#   pip3 install matplotlib


def get_correlations(metric, corr_path):
    frame = metric.frame()
    correlation = frame.corr()
    with open(corr_path, "w") as corr_file:
        corr_file.write(correlation.to_csv())

    all_pairs = {}
    # print(dir(correlation))
    for key in correlation:
        row = correlation[key]
        # print("row dir:", dir(row), list(row.keys()), list(row))
        # print("type: {}, key: {}, row: {}\n\n".format(type(key), key, row))
        for (to, val) in zip(row.keys(), list(row)):
            if key == to or (to, key) in all_pairs:
                continue
            all_pairs[(key, to)] = val

    sorted_pairs = sorted([x for x in all_pairs.items() if not math.isnan(x[1])], key=lambda x: x[1])
    for ((s1, s2), val) in sorted_pairs[:10]:
        print("{:.3f}: {} to {}".format(val, s1, s2))
    print("\n\n")
    for ((s1, s2), val) in sorted_pairs[-10:]:
        print("{:.3f}: {} to {}".format(val, s1, s2))
    return correlation


def combine_correlations(correlations, service_pairs):
    # keys = set().union(*(list(c.keys()) for c in correlations.values()))
    # print(keys)
    # combined = {}
    # for fr in keys:
    #     for to in keys:
    #         if to >= fr:
    #             continue
    #         for (ns, corr) in correlations.items():
    #             print("{:.3f}: {} to {}, {}".format(corr.get(fr, {}).get(to, 0), fr, to, ns))
    #         print("\n")
    #     print("\n")
    for (fr, to) in service_pairs:
        for (ns, corr) in correlations.items():
            print("{:.3f}: {} to {}, {}".format(corr.get(fr, {}).get(to, 0), fr, to, ns))
        print("\n")


def combine_prometheus_new_relic(prom_metrics, new_relic_data):
    print("got {} metrics, {} nr".format(len(prom_metrics), len(new_relic_data)))
    c = collections.defaultdict(lambda: {})
    for (ts, val) in new_relic_data:
        c[ts]["nr"] = val
    for p in prom_metrics:
        auth = p.select("container_name", "auth-server")
        # print(auth.series)
        # print(p.get_key_values(["auth-server", "timestamps"]).frame())
        print("keys: {}".format(auth.series.keys()))
        for (ts, val) in zip(auth.series[()]['timestamps'], auth.series[()]["values"]):
            # print("with seconds: {}".format(int(ts)))
            no_seconds = int(int(ts) / 60) * 60
            # print("without seconds: {}".format(no_seconds))
            cleaned = datetime.utcfromtimestamp(no_seconds)
            c[cleaned]["prom"] = val
    xs, nr, pr, rat = [], [], [], []
    for (ts, v) in sorted(c.items(), key=lambda x: x[0]):
        xs.append(ts)
        nr_val = v.get("nr", 0)
        nr.append(nr_val)
        prom_val = v.get("prom", 0)
        pr.append(prom_val)
        ratio = 0 if nr_val == 0 else prom_val / nr_val
        rat.append(ratio)
        print(ts, nr_val, prom_val, ratio)

    print(list(map(len, [xs, nr, pr, rat])))

    print(xs)
    print(nr)
    plt.plot(xs, nr)
    plt.show()
    plt.plot(xs, pr)
    plt.show()
    plt.plot(xs, rat)
    plt.show()


def combine_nr_prom_example(prom, namespace):
    start = None
    prom_metrics = []
    grab_fresh_data = False
    if grab_fresh_data:
        for i in range(0, 6):
            if start is None:
                start, end = prometheus.interval(hours=4)
            else:
                start, end = prometheus.interval(end=start, hours=4)
            print(start, start.timestamp())
            print(end, end.timestamp())
            print("\n\n")
            path = "cpu_ingress_10_minutes/prom_data/{}-{}.json".format(int(start.timestamp()), int(end.timestamp()))
            prom.get_cpu_by(namespace, path, start, end, 60)
            prom_metrics.append(prometheus.Metric.read_file(path))
    else:
        for path in os.listdir("cpu_ingress_10_minutes/prom_data"):
            prom_metrics.append(prometheus.Metric.read_file(os.path.join("cpu_ingress_10_minutes/prom_data", path)))

    nr_path = "cpu_ingress_10_minutes/{}_auth-server.csv".format(namespace)
    combine_prometheus_new_relic(prom_metrics, new_relic.parse_new_relic(nr_path))


def combine_prometheus_new_relic_cpu_seconds(prom_metrics, new_relic_data):
    print("got {} metrics, {} nr".format(len(prom_metrics), len(new_relic_data)))
    c = collections.defaultdict(lambda: {})
    for (ts, val) in new_relic_data:
        c[ts]["nr"] = val
    for p in prom_metrics:
        auth = p.select("container_name", "auth-server")
        # print(auth.series)
        # print(p.get_key_values(["auth-server", "timestamps"]).frame())
        print("keys: {}".format(auth.series.keys()))
        for (ts, val) in zip(auth.series[()]['timestamps'], auth.series[()]["values"]):
            # print("with seconds: {}".format(int(ts)))
            no_seconds = int(int(ts) / 900) * 900
            # print("without seconds: {}".format(no_seconds))
            cleaned = datetime.utcfromtimestamp(no_seconds)
            c[cleaned]["prom"] = val
    xs, nr, pr, rat, inv_rat = [], [], [], [], []
    prev = None
    for (ts, v) in sorted(c.items(), key=lambda x: x[0]):
        prom_val = v.get("prom")
        if prom_val is None:
            print("skipping 1: {}".format((ts, v)))
            continue
        if prev is None:
            print("skipping 2: {}".format((ts, v)))
            prev = prom_val
            continue

        diff = prom_val - prev
        prev = prom_val

        xs.append(ts)
        nr_val = v.get("nr", 0)
        nr.append(nr_val)
        pr.append(diff)
        ratio = 0 if nr_val == 0 else diff / nr_val
        rat.append(ratio)
        inv = 0 if diff == 0 else nr_val / diff
        inv_rat.append(inv)
        print(ts, nr_val, prom_val, diff, ratio, inv)

    print(list(map(len, [xs, nr, pr, rat])))

    print(xs)
    print(nr)
    home = os.path.expanduser("~")
    plt.plot(xs, nr)
    plt.savefig(os.path.join(home, 'requests.png'))
    plt.show()

    plt.plot(xs, pr)
    plt.savefig(os.path.join(home, 'cpu.png'))
    plt.show()

    plt.plot(xs, rat)
    plt.savefig(os.path.join(home, 'cpu_over_requests.png'))
    plt.show()

    plt.plot(xs, inv_rat)
    plt.savefig(os.path.join(home, 'requests_over_cpu.png'))
    plt.show()


def combine_nr_prom_cpu_usage_ingress(prom, namespace):
    start = None
    prom_metrics = []
    grab_fresh_data = False
    if grab_fresh_data:
        for i in range(0, 6):
            if start is None:
                start, end = prometheus.interval(hours=4)
            else:
                start, end = prometheus.interval(end=start, hours=4)
            print(start, start.timestamp())
            print(end, end.timestamp())
            print("\n\n")
            path = "cpu_ingress_10_minutes/prom_data/{}-{}.json".format(int(start.timestamp()), int(end.timestamp()))
            prom.get_cpu_seconds(namespace, path, start, end, 15 * 60)
            prom_metrics.append(prometheus.Metric.read_file(path))
    else:
        for path in os.listdir("cpu_ingress_10_minutes/prom_data"):
            prom_metrics.append(prometheus.Metric.read_file(os.path.join("cpu_ingress_10_minutes/prom_data", path)))

    nr_path = "cpu_ingress_10_minutes/{}_auth-server.csv".format(namespace)
    combine_prometheus_new_relic_cpu_seconds(prom_metrics, new_relic.parse_new_relic(nr_path, bucket_seconds=900))


if __name__ == "__main__":
    url, namespace = sys.argv[1], sys.argv[2]
    prom = prometheus.Prometheus(url)
    # prom.get_cpu("all_cpu_more_steps.json", "2419") # 345)
    # cpu = Metric.read_file("all_cpu.json")
    # print(cpu.keys)
    # prom.get_ingress(namespace, "ingress.json")
    # prom.get_cpu(namespace, "cpu_full_day.json")
    # get_correlations("ingress.json", "ingress_correlation.csv", "service")
    # get_correlations("cpu_full_day.json", "cpu_correlation_full_day.csv", "container_name")

    # combine_nr_prom_example(prom, namespace)
    combine_nr_prom_cpu_usage_ingress(prom, namespace)

    # correlations = {}
    # for ns in cpu.get_key_values("namespace"):
    #     print("namespace {}".format(ns))
    #     correlations[ns] = get_correlations(cpu.select("namespace", ns), "correlations/cpu_full_week_{}.csv".format(ns))
    #     print("\n\n")
    # combine_correlations(
    #     dict((k, v) for (k, v) in correlations.items() if k not in ("kube-system", "monitoring")),
    #     [("issue-server", "auth-server"), ("issue-server", "rp-polaris-agent-service")]
    # )
