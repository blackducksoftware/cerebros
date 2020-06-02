import json
import subprocess
import sys


def build_hubscan_config(scan_type, version, scan, image_repo, image_sha, image_tag, hub_url, hub_username, hub_password):
    project = image_repo
    return {
        "HubURL": hub_url,
        "HubUsername": hub_username,
        "HubPassword": hub_password,
        "HubPort": 443,

        "HubProjectName": project,
        "HubProjectVersionName": version,
        "HubScanName": scan,

        "ImageRepository": image_repo,
        "ImageSha": image_sha,

        "ImageDirectory": "/tmp/images",

        "ImageTag": image_tag,

        "ScanType": scan_type,

        "LogLevel": "debug"
    }


def build_iscan_config(**args):
    return build_hubscan_config("iscan", "iscan", scan = "{}-iscan".format(args["image_sha"][:10]), **args)


def build_detect_config(tool, **args):
    return build_hubscan_config(tool, tool, scan = "{}-{}".format(args["image_sha"][:10], tool), **args)


def build_configs(**args):
    configs = []
    for tool in ["docker", "signature", "binary"]:
        configs.append(build_detect_config(tool, **args))
    configs.append(build_iscan_config(**args))
    return configs


conf = {
    "hub_url": "",
    "hub_username": "",
    "hub_password": "",

   "image_repo": "docker.io/alpine",
   "image_sha": "ab00606a42621fb68f2ed6ad3c88be54397f981a7b70a79db3d1172b11c4367d",
   "image_tag": "alpine:latest"

#     "image_repo": "library/busybox",
#     "image_sha": "6915be4043561d64e0ab0f8f098dc2ac48e077fe23f488ac24b665166898115a",
#     "image_tag": "busybox:latest"

#    "image_repo": "gcr.io/distroless/java-debian9",
#    "image_sha": "b715126ebd36e5d5c2fd730f46a5b3c3b760e82dc18dffff7f5498d0151137c9",
#    "image_tag": "gcr.io/distroless/java-debian9:11"
}

# with open(sys.argv[1]) as conf_file:
#    conf = json.load(conf_file)

for c in build_configs(**conf):
    proj = c["HubProjectName"].replace("/", "-")
    conf_path = "{}-{}.json".format(proj, c["HubProjectVersionName"])
    with open(conf_path, "w") as new_conf:
        json_dump = json.dumps(c, indent=2)
        new_conf.write(json_dump)
    command = ["go", "run", "blackduck-cli-single-scan.go", conf_path]
    print("about to run: <{}> with config {}".format(" ".join(command), json_dump))
    proc = subprocess.Popen(command,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE)
    stdout, stderr = proc.communicate()
    print("stdout and stderr: \n{}\n{}\n".format(stdout, stderr))
