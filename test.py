"""
Roamify - Grok AI Feature Test
Tests the 3 AI-powered endpoints to confirm GROK_KEY is working on Render.

Usage:
    python test_grok.py
"""

import requests
import json
import time
from datetime import datetime, timedelta, timezone

BASE = "https://roamify-f9v5.onrender.com"
unique_email = f"groktest_{int(time.time())}@example.com"

print("=" * 60)
print("STEP 1: Authenticating...")
print("=" * 60)

r = requests.post(f"{BASE}/api/v1/auth/register", json={
    "name": "GrokTester",
    "email": unique_email,
    "password": "password123"
}, timeout=20)
print(f"Register ({unique_email}) -> {r.status_code}")

token = None

try:
    data = r.json()
    token = (
        data.get("token") or
        data.get("access_token") or
        data.get("data", {}).get("token") or
        data.get("data", {}).get("access_token")
    )
except Exception:
    pass

if not token:
    r2 = requests.post(f"{BASE}/api/v1/auth/login", json={
        "email": unique_email,
        "password": "password123"
    }, timeout=20)
    print(f"Login -> {r2.status_code}")
    try:
        data = r2.json()
        token = (
            data.get("token") or
            data.get("access_token") or
            data.get("data", {}).get("token") or
            data.get("data", {}).get("access_token")
        )
    except Exception:
        pass

    if not token:
        print("\n❌ Could not get token automatically.")
        print("Register response:", r.text[:300])
        print("Login response:  ", r2.text[:300])
        print()
        print("💡 Go to your Swagger UI, log in manually, and paste the token here:")
        token = input("Paste Bearer token (without 'Bearer '): ").strip()
        if not token:
            exit(1)

print(f"✅ Token obtained: {token[:40]}...")
headers = {"Authorization": f"Bearer {token}"}

MOCK_RESPONSE = "Start central, explore nearby"

# ─────────────────────────────────────────────
# TEST 1: Travel Assistant
# ─────────────────────────────────────────────
print("\n" + "=" * 60)
print("TEST 1: Travel Assistant  POST /api/v1/assistant/travel")
print("=" * 60)

r = requests.post(f"{BASE}/api/v1/assistant/travel", json={
    "prompt": "What are the best things to do in Tokyo for 3 days?",
    "destination": "Tokyo, Japan"
}, headers=headers, timeout=30)
print(f"Status: {r.status_code}")

try:
    body = r.json()
    print(json.dumps(body, indent=2)[:800])
    if MOCK_RESPONSE in json.dumps(body):
        print("\n🔴 RESULT: MOCK FALLBACK — GROK_KEY is still not loaded on Render!")
    elif r.status_code == 200:
        print("\n✅ RESULT: REAL AI RESPONSE — Grok is working!")
    else:
        print(f"\n⚠️  RESULT: Status {r.status_code} — check Render logs")
except Exception:
    print("Raw:", r.text[:500])

# ─────────────────────────────────────────────
# TEST 2: AI Trip Plan & Create
# ─────────────────────────────────────────────
print("\n" + "=" * 60)
print("TEST 2: AI Trip Itinerary  POST /api/v1/trips/plan-and-create")
print("=" * 60)

now   = datetime.now(timezone.utc)
start = (now + timedelta(days=7)).strftime("%Y-%m-%dT%H:%M:%SZ")
end   = (now + timedelta(days=10)).strftime("%Y-%m-%dT%H:%M:%SZ")

r = requests.post(f"{BASE}/api/v1/trips/plan-and-create", json={
    "name": "Tokyo Adventure",
    "location": "Tokyo, Japan",
    "vibe": "cultural",
    "number_of_people": 2,
    "start_date": start,
    "end_date": end,
    "budget": 1500.00,
    "prompt": "Focus on temples, food markets and hidden gems."
}, headers=headers, timeout=45)
print(f"Status: {r.status_code}")

try:
    body = r.json()
    print(json.dumps(body, indent=2)[:800])
    # 200 or 201 both mean success
    if r.status_code in (200, 201):
        itinerary = (
            body.get("itinerary") or
            body.get("data", {}).get("itinerary") or []
        )
        if itinerary:
            titles = [item.get("title", "") for item in itinerary[:3]]
            print(f"\nFirst 3 activity titles: {titles}")
            generic = {"Arrival", "Departure", "Day 1", "Morning", "Afternoon"}
            if any(t in generic for t in titles):
                print("⚠️  RESULT: Generic fallback activities — Grok may not be active")
            else:
                print("✅ RESULT: AI-generated itinerary detected!")
        else:
            print("⚠️  No itinerary array found in response")
    else:
        print(f"⚠️  Status {r.status_code}")
except Exception:
    print("Raw:", r.text[:500])

# ─────────────────────────────────────────────
# TEST 3: Generate Locations
# ─────────────────────────────────────────────
print("\n" + "=" * 60)
print("TEST 3: Generate Locations  POST /api/v1/discovery/locations/generate")
print("=" * 60)

r = requests.post(f"{BASE}/api/v1/discovery/locations/generate", json={
    "answers": {
        "explorer_type": "adventurer",
        "preferred_vibes": ["beach", "cultural"],
        "interests": ["food", "hiking", "history"],
        "budget_style": "budget",
        "travel_pace": "relaxed",
        "travel_with": "partner",
        "desired_region": "Asia",
        "max_budget": 2000.0,
        "trip_length_days": 7
    },
    "limit": 5
}, headers=headers, timeout=30)
print(f"Status: {r.status_code}")

try:
    body = r.json()
    print(json.dumps(body, indent=2)[:600])
    if r.status_code in (200, 201):
        print("✅ RESULT: Endpoint responded successfully!")
    else:
        print(f"⚠️  Status {r.status_code}")
except Exception:
    print("Raw:", r.text[:400])

# ─────────────────────────────────────────────
print("\n" + "=" * 60)
print("SUMMARY")
print("=" * 60)
print("✅ = Grok AI is responding with real data")
print("🔴 = Still mock fallback — GROK_KEY not picked up by Render")
print("⚠️  = Unexpected error — check Render Logs tab for:")
print("     '[startup] GROK_KEY loaded successfully ✅'")
print("=" * 60)