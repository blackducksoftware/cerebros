import collections
from datetime import datetime, timedelta, timezone
import json
import pandas
import requests


def now():
    return datetime.utcnow().replace(tzinfo=timezone.utc).timestamp()

def interval(end=datetime.utcnow().replace(tzinfo=timezone.utc), days=0, hours=0):
    start = (end - timedelta(days=days, hours=hours))
    return start, end


nginx_ingress_controller_requests_namespace_by_service_query = """
sum(
  rate(
    nginx_ingress_controller_requests{{
      namespace="{}"
    }}[5m]
  )
) by (service)
"""

nginx_ingress_controller_requests_query = """
sum(
  rate(nginx_ingress_controller_requests[5m])
) by (namespace, service)
"""

cpu_namespace_by_container_query = """
sum(
  rate(
    container_cpu_usage_seconds_total{{
      namespace="{namespace}",
      image!="",
      container_name!="POD"
    }}[5m]
  )
) by (container_name)

/

sum(
  container_spec_cpu_quota{{
    namespace="{namespace}",
    image!="",
    container_name!="POD"
  }}
  
  /
  
  container_spec_cpu_period{{
    image!="",
    namespace="{namespace}",
    container_name!="POD"
  }}
) by (container_name)
"""

cpu_query = """
sum(
  rate(
    container_cpu_usage_seconds_total{
      image!="",
      container_name!="POD"
    }[5m]
  )
) by (namespace, container_name)

/

sum(
  container_spec_cpu_quota{
    image!="",
    container_name!="POD"
  }
  
  /
  
  container_spec_cpu_period{
    image!="",
    container_name!="POD"
  }
) by (namespace, container_name)
"""

cpu_seconds_query = """
sum(
    container_cpu_usage_seconds_total{{
      image!="",
      namespace="{namespace}",
      container_name!="POD"
    }}
) by (container_name)
"""

class Prometheus:
    def __init__(self, url):
        self.url = url

    def issue_request(self, query, start, end, step):
        url = "{}/api/v1/query_range".format(self.url)
        print("issuing request to {}".format(url))
        params = {'query': query, 'start': start, 'end': end, 'step': step}
        resp = requests.get(url, params=params)
        print("issued request to {}", resp.url)
        return resp

    def get_ingress(self, namespace, path):
        query = nginx_ingress_controller_requests_namespace_by_service_query.format(namespace)
        resp = self.issue_request(
            query,
            "1588853225.194",
            "1588874825.194",
            "86"
        )
        with open(path, "w") as outfile:
            outfile.write(json.dumps(resp.json(), indent=2))
        return resp

    def get_cpu_by(self, namespace, path, start, end, steps=252):
        query = cpu_namespace_by_container_query.format(namespace=namespace)
        resp = self.issue_request(
            query,
            start.timestamp(),
            end.timestamp(),
            steps
        )
        with open(path, "w") as outfile:
            outfile.write(json.dumps(resp.json(), indent=2))
        return resp

    def get_cpu_seconds(self, namespace, path, start, end, steps=252):
        query = cpu_seconds_query.format(namespace=namespace)
        resp = self.issue_request(
            query,
            start.timestamp(),
            end.timestamp(),
            steps
        )
        with open(path, "w") as outfile:
            outfile.write(json.dumps(resp.json(), indent=2))
        return resp

    def get_cpu(self, path, steps):
        start, end = interval(days=7)
        print("start, end: {}, {}".format(start, end))
        resp = self.issue_request(
            cpu_query,
            start.timestamp(),
            end.timestamp(),
            steps)
        with open(path, "w") as outfile:
            outfile.write(json.dumps(resp.json())) # , indent=2))
        return resp


class Metric:
    def __init__(self, keys, series):
        self.keys = dict(zip(sorted(keys), range(len(keys))))
        self.series = series

    # TODO select by multiple keys
    def select(self, key, value):
        ix = self.keys[key]
        series = {}
        for (tuple_key, s) in self.series.items():
            series_key = dict(tuple_key)
            if series_key[key] == value:
                new_key = tuple((k, v) for (k, v) in tuple_key if k != key)
                series[new_key] = s
        remaining_keys = [k for k in self.keys.keys() if k != key]
        return Metric(remaining_keys, series)

    def get_key_values(self, key):
        vals = set()
        for (tuple_key, s) in self.series.items():
            print("tuple_key: {}".format(tuple_key))
            series_key = dict(tuple_key)
            vals.add(series_key[key])
        return vals

    def frame(self, missing_value=0):
        """
        handles series with different numbers of datapoints, by filling in with `missing_value`
        """
        raw_series = {}
        all_timestamps = set()
        for (key, s) in self.series.items():
            # print("found {} for {}".format(len(s["values"]), key))
            all_timestamps = all_timestamps.union(s["timestamps"])
            raw_series["_".join(p[1] for p in key)] = dict(zip(s["timestamps"], s["values"]))
        sorted_timestamps = sorted(all_timestamps)
        cleaned_series = {"timestamps": sorted_timestamps}
        for (key, s) in raw_series.items():
            values = [s.get(ts, missing_value) for ts in sorted_timestamps]
            cleaned_series[key] = values
        return pandas.DataFrame(cleaned_series)

    @staticmethod
    def read_file(path):
        with open(path) as infile:
            blob = json.load(infile)
            if blob["status"] != "success":
                raise ValueError("can't read status {}".format(blob["status"]))
        series = {}
        keys = collections.Counter()
        for metric in blob["data"]["result"]:
            key = tuple((k, v) for (k, v) in sorted(metric["metric"].items(), key=lambda x: x[0]))
            keys[tuple(metric["metric"].keys())] += 1
            series[key] = {
                "values": [float(p[1]) for p in metric["values"]],
                # TODO parse these into datetimes or something
                "timestamps": [p[0] for p in metric["values"]]
            }
        if len(keys) != 1:
            raise ValueError("expected all keys to match, got {}".format(str(keys)))
        return Metric(list(keys.keys())[0], series)
