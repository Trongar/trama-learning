#!/usr/bin/env python3
"""End-to-end PocketBase account and record isolation check for localhost only."""

import json
import os
import secrets
import urllib.error
import urllib.parse
import urllib.request
import uuid
from typing import Any

BASE = os.environ.get("TRAMA_TEST_URL", "http://127.0.0.1:8091").rstrip("/")
if urllib.parse.urlparse(BASE).hostname not in {"127.0.0.1", "localhost", "::1"}:
    raise SystemExit("Refusing to create test accounts outside localhost.")
PASSWORD = "Trama-test-" + secrets.token_urlsafe(24)


def request(path, method="GET", body=None, token=None) -> tuple[int, dict[str, Any]]:
    headers = {"Accept": "application/json"}
    data = None
    if body is not None:
        headers["Content-Type"] = "application/json"
        data = json.dumps(body).encode()
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(BASE + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=15) as response:
            payload = response.read()
            decoded = json.loads(payload) if payload else {}
            return response.status, decoded if isinstance(decoded, dict) else {}
    except urllib.error.HTTPError as error:
        payload = error.read()
        try:
            decoded = json.loads(payload) if payload else {}
        except json.JSONDecodeError:
            decoded = {}
        return error.code, decoded if isinstance(decoded, dict) else {}


def expect(condition, description):
    if not condition:
        raise AssertionError(description)
    print("PASS:", description)


accounts = []
for _ in range(2):
    email = f"trama-{uuid.uuid4().hex}@example.com"
    status, _ = request("/api/collections/users/records", "POST", {
        "email": email,
        "password": PASSWORD,
        "passwordConfirm": PASSWORD,
        "emailVisibility": False,
    })
    expect(status in (200, 201), f"account signup returns success (HTTP {status})")
    status, auth = request("/api/collections/users/auth-with-password", "POST", {
        "identity": email,
        "password": PASSWORD,
    })
    expect(status == 200 and bool(auth.get("token")), "PocketBase issues an auth token")
    accounts.append({"id": auth["record"]["id"], "token": auth["token"]})

lists = []
for account in accounts:
    status, result = request("/api/collections/learning_goals/records?sort=-created&perPage=100", token=account["token"])
    expect(status == 200, "authenticated account can list its own learning routes")
    lists.append(result["items"])
expect(len(lists[0]) == 1 and len(lists[1]) == 1, "each new account receives exactly one preconfigured route")
goal_a = lists[0][0]
goal_b = lists[1][0]
expect(goal_a["id"] != goal_b["id"] and goal_a["user"] == accounts[0]["id"] and goal_b["user"] == accounts[1]["id"], "starter routes are separate and correctly owned")
expect(len(goal_a["path"]) == 3 and len(goal_b["path"]) == 3, "each initial route contains three learning steps")

status, _ = request("/api/collections/users/records/" + accounts[0]["id"], token=accounts[1]["token"])
expect(status == 404, "another account cannot view the first account's profile")
status, users_b = request("/api/collections/users/records?perPage=100", token=accounts[1]["token"])
expect(status == 200 and len(users_b["items"]) == 1 and users_b["items"][0]["id"] == accounts[1]["id"], "account listing exposes only the signed-in profile")

status, _ = request("/api/collections/learning_goals/records/" + goal_a["id"], token=accounts[1]["token"])
expect(status == 404, "another account cannot view a route by record id")
status, _ = request("/api/trama/goals/" + goal_a["id"] + "/attempts", "POST", {"index": 0, "answer": "Private test response"}, accounts[1]["token"])
expect(status == 404, "another account cannot submit attempts to that route")
status, result = request("/api/trama/goals/" + goal_a["id"] + "/attempts", "POST", {"index": 0, "answer": "Aprender a aprender importa para mi objetivo"}, accounts[0]["token"])
expect(status == 200 and "assessment" in result, "owner can submit an answer and get demo feedback")
status, refreshed = request("/api/collections/learning_goals/records?perPage=100", token=accounts[0]["token"])
expect(status == 200 and float(refreshed["items"][0]["progress"]) > 0, "owner sees saved progress after an attempt")
status, listing_b = request("/api/collections/learning_goals/records?perPage=100", token=accounts[1]["token"])
expect(status == 200 and all(item["id"] != goal_a["id"] for item in listing_b["items"]), "the other account's listing remains isolated")
status, _ = request("/api/collections/learning_goals/records", "POST", {
    "user": accounts[0]["id"],
    "domain": "Cross-account injection",
    "level": "inicial",
    "purpose": "must be denied",
    "minutes_per_week": 90,
    "status": "active",
    "progress": 0,
    "path": [{"name": "probe", "summary": "probe", "content": "probe", "prompt": "probe", "mastery": 0}],
}, accounts[1]["token"])
expect(status != 200, "PocketBase rejects creating a route for another account")
status, _ = request("/api/collections/learning_goals/records/" + goal_a["id"], "PATCH", {"user": accounts[1]["id"]}, accounts[0]["token"])
expect(status != 200, "PocketBase prevents transferring route ownership")
print("PASS: PocketBase account isolation integration checks completed")
