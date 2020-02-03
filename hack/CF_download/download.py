import pandas as pd
import os
import json
import requests

from google.cloud import pubsub_v1

API_ENDPOINT = "https://XXXX.cloudfunctions.net/jdtest2"
headers = {'content-type': 'application/json'}

os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "/Users/jeremyd/Downloads/XXXX"

df = pd.read_csv("data-1580396677735.csv",
                 na_values="?", comment='\t',
                 sep=",", skipinitialspace=True, low_memory=False)

print(df.isna().sum())

project_id = "XXXX"
topic_name = "XXXX"

publisher = pubsub_v1.PublisherClient()
topic_path = publisher.topic_path(project_id, topic_name)

# Only scan 10000 projects
df = df.truncate(before=None, after=10000)

for index, row in df.iterrows():
    name = row['forge_id'].replace("/", "-")

    data = {'url': "http://github.com/" + row['forge_id'] + "/archive/master.zip",
            'name': name}

    data = str(json.dumps(data)).encode("utf-8")
    future = publisher.publish(topic_path, data=data)
    print(future.result())