#! python3
import csv
import json


def read_projects(path="kb_github_projects.csv"):
    projects = []
    with open(path) as gh_file:
        reader = csv.reader(gh_file, delimiter=",")
        headers = next(reader)
        for row in reader:
            projects.append(dict(zip(headers, row)))
    return projects


projs = read_projects()

js_projs = [p for p in sorted(projs, key=lambda r: -1 * r["size"]) if p["language"] == "JavaScript"]

print(json.dumps(js_projs[:5], indent=2))

#print("\n".join("git clone git@github.com:{}".format(p["forge_id"]) for p in js_projs[:50]))
out = {}
for p in js_projs[:50]:
    org, repo = p["forge_id"].split("/")
#     print(json.dumps(p, indent=2))
#     print(json.dumps()
    out[repo] = {"Repo": p["forge_id"]}
print(json.dumps(out, indent=2))
