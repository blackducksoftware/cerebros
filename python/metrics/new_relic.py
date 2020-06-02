import csv
import datetime
import matplotlib.pyplot as plt


def parse_new_relic(path, bucket_seconds=60):
    pairs = []
    with open(path) as infile:
        reader = csv.reader(infile)
        for (i, row) in enumerate(reader):
            if i == 0:
                print("skipping header: {}".format(row))
                # print("<>".join(row))
                continue
            _, _, stamp, value = row
            ts = datetime.datetime.utcfromtimestamp((int(stamp) * bucket_seconds * 1000) / 1e3)
            # print(ts.timestamp())
            pairs.append((ts, int(value)))
    return sorted(pairs, key=lambda p: p[0])


def example_plot(path):
    vals = parse_new_relic(path)
    for v in vals[:10]:
        print(v[0].timestamp(), v)

    x, y = zip(*vals)
    plt.plot(x, y)

    plt.show()


if __name__ == "__main__":
    import sys
    path = sys.argv[1]
    example_plot(path)
