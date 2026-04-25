import requests
import json
import time
import sys
from datetime import datetime
from typing import Optional, Dict, Any

BASE_URL = "https://roamify-f9v5.onrender.com"


def safe_json(response: requests.Response) -> Any:
    try:
        return response.json()
    except ValueError:
        return {"text": response.text}


def extract_token(body: Any) -> Optional[str]:
    if isinstance(body, dict):
        data = body.get("data", {})
        if isinstance(data, dict):
            return data.get("token") or data.get("access_token") or data.get("jwt")
    return None


def extract_id(body: Any) -> Optional[str]:
    if isinstance(body, dict):
        data = body.get("data", {})
        if isinstance(data, dict):
            if "id" in data:
                return data["id"]
        if "id" in body:
            return body["id"]
    return None


class TestRunner:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.session = requests.Session()
        self.user1_email = None
        self.user2_email = None
        self.user1_token = None
        self.user2_token = None
        self.user1_id = None
        self.user2_id = None
        self.trip_id = None
        self.post_id = None
        self.comment_id = None
        self.wishlist_item_id = None
        self.wishlist_collection_id = None
        self.challenge_id = None
        self.trivia_question_id = None
        self.report_id = None
        self.test_count = 0
        self.passed_count = 0

    def make_request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict] = None,
        auth_token: Optional[str] = None,
        expect_status: Optional[int] = None,
        params: Optional[Dict] = None,
    ) -> Dict:
        url = self.base_url + endpoint
        headers = {}
        if auth_token:
            headers["Authorization"] = f"Bearer {auth_token}"

        try:
            if method in ("GET", "DELETE"):
                response = self.session.request(
                    method, url, headers=headers, params=params, timeout=15
                )
            else:
                response = self.session.request(
                    method, url, headers=headers, json=data or {}, params=params, timeout=15
                )

            self.test_count += 1
            body = safe_json(response)
            status = response.status_code
            success = (
                status == expect_status
                if expect_status is not None
                else 200 <= status < 300
            )
            if success:
                self.passed_count += 1

            print(
                f"{'PASS' if success else 'FAIL'} {method:6s} {endpoint:50s} -> {status}"
            )
            print(f"  Body: {json.dumps(body)[:320]}")
            return {"status": status, "success": success, "body": body}
        except Exception as exc:
            self.test_count += 1
            print(f"ERROR {method:6s} {endpoint:50s} -> {exc}")
            return {"status": 0, "success": False, "body": {"error": str(exc)}}

    def run_all_tests(self) -> bool:
        print("=" * 80)
        print("ROAMIFY API TEST SUITE")
        print(f"Base URL: {self.base_url}")
        print(f"Timestamp: {datetime.now().isoformat()}")
        print("=" * 80)

        self.test_health()
        self.test_swagger()
        self.test_register_users()
        self.test_login_users()
        self.test_user_endpoints()
        self.test_trips()
        self.test_posts()
        self.test_wishlist()
        self.test_challenges()
        self.test_passport()
        self.test_discovery()
        self.test_reports()
        self.test_search()

        print("=" * 80)
        print(f"RESULTS: {self.passed_count}/{self.test_count} passed")
        print("=" * 80)
        return self.passed_count == self.test_count

    def test_health(self):
        print("\n--- PUBLIC HEALTH ---")
        self.make_request("GET", "/health")

    def test_swagger(self):
        print("\n--- SWAGGER ---")
        self.make_request("GET", "/swagger")
        self.make_request("GET", "/swagger/index.html")

    def test_register_users(self):
        print("\n--- REGISTER USERS ---")
        timestamp = int(time.time() * 1000)
        self.user1_email = f"testuser1_{timestamp}@example.com"
        self.user2_email = f"testuser2_{timestamp}@example.com"

        result = self.make_request(
            "POST",
            "/api/v1/auth/register",
            {"full_name": "Test User 1", "email": self.user1_email, "password": "TestPassword123!"},
            expect_status=201,
        )
        self.user1_id = extract_id(result["body"])

        result = self.make_request(
            "POST",
            "/api/v1/auth/register",
            {"full_name": "Test User 2", "email": self.user2_email, "password": "TestPassword123!"},
            expect_status=201,
        )
        self.user2_id = extract_id(result["body"])

    def test_login_users(self):
        print("\n--- LOGIN USERS ---")
        if not self.user1_email or not self.user2_email:
            print("Skipping login because registration failed.")
            return

        result = self.make_request(
            "POST",
            "/api/v1/auth/login",
            {"email": self.user1_email, "password": "TestPassword123!"},
        )
        self.user1_token = extract_token(result["body"])

        result = self.make_request(
            "POST",
            "/api/v1/auth/login",
            {"email": self.user2_email, "password": "TestPassword123!"},
        )
        self.user2_token = extract_token(result["body"])

        if not self.user1_id and self.user1_token:
            me_result = self.make_request("GET", "/api/v1/users/me", auth_token=self.user1_token)
            if me_result["success"]:
                self.user1_id = extract_id(me_result["body"])

    def test_user_endpoints(self):
        print("\n--- USER ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping user endpoints because auth failed.")
            return

        self.make_request("GET", "/api/v1/users/me", auth_token=self.user1_token)
        self.make_request(
            "PATCH",
            "/api/v1/users/me",
            {"full_name": "Test User 1 Updated"},
            auth_token=self.user1_token,
        )
        self.make_request("GET", "/api/v1/users/me/vibe", auth_token=self.user1_token)
        self.make_request(
            "PUT",
            "/api/v1/users/me/vibe",
            {"preferred_vibes": ["adventure"], "travel_pace": "fast"},
            auth_token=self.user1_token,
        )
        self.make_request("GET", "/api/v1/users/me/privacy", auth_token=self.user1_token)
        self.make_request(
            "PATCH",
            "/api/v1/users/me/privacy",
            {"ghost_mode_enabled": False, "map_visibility": "public"},
            auth_token=self.user1_token,
        )
        self.make_request("GET", "/api/v1/users/search?q=test", auth_token=self.user1_token)

        if self.user2_id:
            self.make_request(
                "POST",
                "/api/v1/users/follow",
                {"user_id": str(self.user2_id)},
                auth_token=self.user1_token,
            )
            self.make_request(
                "DELETE", f"/api/v1/users/follow/{self.user2_id}", auth_token=self.user1_token
            )

        if self.user1_id:
            self.make_request(f"GET", f"/api/v1/users/{self.user1_id}", auth_token=self.user1_token)
            self.make_request(
                "GET", f"/api/v1/users/{self.user1_id}/followers", auth_token=self.user1_token
            )
            self.make_request(
                "GET", f"/api/v1/users/{self.user1_id}/following", auth_token=self.user1_token
            )

    def test_trips(self):
        print("\n--- TRIPS ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping trips tests because auth failed.")
            return

        trip_data = {
            "title": f"Paris Adventure {int(time.time())}",
            "destination": "Paris, France",
            "start_date": "2025-06-01T00:00:00Z",
            "end_date": "2025-06-08T00:00:00Z",
            "budget": 2500,
        }
        result = self.make_request("POST", "/api/v1/trips", trip_data, auth_token=self.user1_token, expect_status=201)
        self.trip_id = extract_id(result["body"])

        if self.trip_id:
            self.make_request("GET", "/api/v1/trips", auth_token=self.user1_token)
            self.make_request("GET", f"/api/v1/trips/{self.trip_id}", auth_token=self.user1_token)
            self.make_request(
                "PATCH",
                f"/api/v1/trips/{self.trip_id}",
                {"title": "Paris Adventure Updated"},
                auth_token=self.user1_token,
            )

            if self.user2_id:
                self.make_request(
                    "POST",
                    f"/api/v1/trips/{self.trip_id}/members",
                    {"user_id": str(self.user2_id)},
                    auth_token=self.user1_token,
                    expect_status=201,
                )
                self.make_request("GET", f"/api/v1/trips/{self.trip_id}/members", auth_token=self.user1_token)

            self.make_request("GET", f"/api/v1/trips/{self.trip_id}/map", auth_token=self.user1_token)

    def test_posts(self):
        print("\n--- POSTS ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping posts tests because auth failed.")
            return

        post_data = {"content": "Test post content from endpoint_test.py", "visibility": "public"}
        result = self.make_request("POST", "/api/v1/posts", post_data, auth_token=self.user1_token, expect_status=201)
        self.post_id = extract_id(result["body"])

        if self.post_id:
            self.make_request("GET", "/api/v1/posts", auth_token=self.user1_token)
            self.make_request("GET", f"/api/v1/posts/{self.post_id}", auth_token=self.user1_token)
            self.make_request(
                "PATCH",
                f"/api/v1/posts/{self.post_id}",
                {"content": "Updated content for the test post"},
                auth_token=self.user1_token,
            )
            self.make_request("POST", f"/api/v1/posts/{self.post_id}/like", auth_token=self.user1_token)
            self.make_request("DELETE", f"/api/v1/posts/{self.post_id}/like", auth_token=self.user1_token)

            comment_result = self.make_request(
                "POST",
                f"/api/v1/posts/{self.post_id}/comments",
                {"content": "Nice post!"},
                auth_token=self.user1_token,
                expect_status=201,
            )
            self.comment_id = extract_id(comment_result["body"])

            self.make_request("GET", f"/api/v1/posts/{self.post_id}/comments", auth_token=self.user1_token)
            if self.comment_id:
                self.make_request(
                    "DELETE",
                    f"/api/v1/posts/{self.post_id}/comments",
                    auth_token=self.user1_token,
                    params={"comment_id": str(self.comment_id)},
                )

            if self.user1_id:
                self.make_request(f"GET", f"/api/v1/users/{self.user1_id}/posts", auth_token=self.user1_token)

    def test_wishlist(self):
        print("\n--- WISHLIST ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping wishlist tests because auth failed.")
            return

        item_result = self.make_request(
            "POST",
            "/api/v1/wishlist/items",
            {"name": "Test wishlist item"},
            auth_token=self.user1_token,
            expect_status=201,
        )
        self.wishlist_item_id = extract_id(item_result["body"])

        self.make_request("GET", "/api/v1/wishlist/items", auth_token=self.user1_token)
        if self.wishlist_item_id:
            self.make_request(
                "PATCH",
                f"/api/v1/wishlist/items/{self.wishlist_item_id}",
                {"name": "Updated wishlist item"},
                auth_token=self.user1_token,
            )

        collection_result = self.make_request(
            "POST",
            "/api/v1/wishlist/collections",
            {"name": "Test collection"},
            auth_token=self.user1_token,
            expect_status=201,
        )
        self.wishlist_collection_id = extract_id(collection_result["body"])

        self.make_request("GET", "/api/v1/wishlist/collections", auth_token=self.user1_token)
        if self.wishlist_collection_id:
            self.make_request(
                "GET",
                f"/api/v1/wishlist/collections/{self.wishlist_collection_id}",
                auth_token=self.user1_token,
            )
            if self.wishlist_item_id:
                self.make_request(
                    "POST",
                    f"/api/v1/wishlist/collections/{self.wishlist_collection_id}/items",
                    {"wishlist_item_id": str(self.wishlist_item_id)},
                    auth_token=self.user1_token,
                )

    def test_challenges(self):
        print("\n--- CHALLENGES ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping challenges tests because auth failed.")
            return

        self.make_request("GET", "/api/v1/challenges", auth_token=self.user1_token)
        self.make_request("GET", "/api/v1/challenges/leaderboard", auth_token=self.user1_token)

        challenge_result = self.make_request(
            "POST",
            "/api/v1/challenges",
            {"title": "Visit 5 Museums", "description": "Challenge test"},
            auth_token=self.user1_token,
            expect_status=201,
        )
        self.challenge_id = extract_id(challenge_result["body"])

        self.make_request("GET", "/api/v1/challenges/my-progress", auth_token=self.user1_token)
        self.make_request("GET", "/api/v1/challenges/trivia", auth_token=self.user1_token)

        trivia_result = self.make_request(
            "POST",
            "/api/v1/challenges/trivia",
            {"question": "What is the capital of France?", "choices": ["Paris", "London", "Berlin"], "correct_answer": "Paris"},
            auth_token=self.user1_token,
            expect_status=201,
        )
        self.trivia_question_id = extract_id(trivia_result["body"])

        if self.challenge_id:
            self.make_request(
                "POST",
                "/api/v1/challenges/accept",
                {"challenge_id": str(self.challenge_id)},
                auth_token=self.user1_token,
                expect_status=201,
            )
            self.make_request(
                "POST",
                "/api/v1/challenges/complete",
                {"challenge_id": str(self.challenge_id)},
                auth_token=self.user1_token,
            )

        if self.trivia_question_id:
            self.make_request(
                "POST",
                "/api/v1/challenges/trivia/answer",
                {"question_id": str(self.trivia_question_id), "answer": "Paris"},
                auth_token=self.user1_token,
            )

    def test_passport(self):
        print("\n--- PASSPORT ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping passport tests because auth failed.")
            return

        self.make_request(
            "PUT",
            "/api/v1/passport/vault",
            {"encrypted_payload": "dGVzdCBwYXNzZXBvcnQ="},
            auth_token=self.user1_token,
        )
        self.make_request("GET", "/api/v1/passport/vault", auth_token=self.user1_token)

        stamp_result = self.make_request(
            "POST",
            "/api/v1/passport/stamps",
            {"country": "France", "country_code": "FR", "date_visited": "2025-05-20T00:00:00Z"},
            auth_token=self.user1_token,
            expect_status=201,
        )
        if extract_id(stamp_result["body"]):
            print(f"  Created passport stamp id: {extract_id(stamp_result['body'])}")

        self.make_request("GET", "/api/v1/passport/stamps", auth_token=self.user1_token)

    def test_discovery(self):
        print("\n--- DISCOVERY ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping discovery tests because auth failed.")
            return

        self.make_request("GET", "/api/v1/discovery/home", auth_token=self.user1_token)
        self.make_request("GET", "/api/v1/discovery/vibe-search", auth_token=self.user1_token, params={"q": "adventure"})
        self.make_request("GET", "/api/v1/discovery/atlas", auth_token=self.user1_token)
        self.make_request("GET", "/api/v1/discovery/atlas/geojson", auth_token=self.user1_token)
        self.make_request("GET", "/api/v1/discovery/price-drops", auth_token=self.user1_token)
        self.make_request("GET", "/api/v1/discovery/recommended", auth_token=self.user1_token)
        self.make_request("GET", "/api/v1/home", auth_token=self.user1_token)
        self.make_request("GET", "/api/v1/vibe-search", auth_token=self.user1_token, params={"q": "city"})
        self.make_request("GET", "/api/v1/atlas", auth_token=self.user1_token)

    def test_reports(self):
        print("\n--- REPORTS ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping reports tests because auth failed.")
            return

        if not self.user2_id:
            print("Skipping report creation because no second user id is available.")
            return

        self.make_request(
            "POST",
            "/api/v1/reports",
            {"target_type": "user", "target_id": str(self.user2_id), "reason": "inappropriate"},
            auth_token=self.user1_token,
            expect_status=201,
        )

    def test_search(self):
        print("\n--- SEARCH ENDPOINTS ---")
        if not self.user1_token:
            print("Skipping search tests because auth failed.")
            return

        self.make_request("GET", "/api/v1/search/global", auth_token=self.user1_token, params={"q": "paris"})
        self.make_request("GET", "/api/v1/search/results", auth_token=self.user1_token, params={"q": "travel"})


def main():
    runner = TestRunner(BASE_URL)
    passed = runner.run_all_tests()
    sys.exit(0 if passed else 1)


if __name__ == "__main__":
    main()