#!/usr/bin/env python3
# Simple script to send test webhooks
import sys
import requests

if len(sys.argv) != 2:
    print("Provide name of webhook as argument")
    sys.exit(1)

webhook_type = sys.argv[1]

try:
    with open("./{}.json".format(webhook_type), "r") as f:
        json = f.read()
except FileNotFoundError:
    print("Can't find {}.json".format(webhook_type))
    sys.exit(1)

url = "http://localhost:8045/github"
print("Posting to {}...".format(url))

res = requests.post(url, data=json, headers={"X-GitHub-Event": webhook_type})

if res.status_code != 200:
    print("Webhook error (status {}):".format(res.status_code))
    print(res.text)
    sys.exit(2)
else:
    print("Success!")
