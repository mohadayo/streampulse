import pytest
from app import app


@pytest.fixture
def client():
    app.config["TESTING"] = True
    with app.test_client() as client:
        yield client


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "healthy"
    assert data["service"] == "event-collector"
    assert "timestamp" in data


def test_collect_event(client):
    resp = client.post("/events", json={"event_type": "page_view", "payload": {"url": "/home"}, "source": "web"})
    assert resp.status_code == 201
    data = resp.get_json()
    assert data["message"] == "Event collected"
    assert "event_id" in data


def test_collect_event_missing_body(client):
    resp = client.post("/events", content_type="application/json", data="")
    assert resp.status_code == 400


def test_collect_event_missing_event_type(client):
    resp = client.post("/events", json={"payload": {"key": "val"}})
    assert resp.status_code == 400
    data = resp.get_json()
    assert "event_type" in data["error"]


def test_list_events(client):
    client.post("/events", json={"event_type": "click"})
    client.post("/events", json={"event_type": "scroll"})
    resp = client.get("/events")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["total"] >= 2


def test_list_events_filter(client):
    client.post("/events", json={"event_type": "filter_test"})
    resp = client.get("/events?event_type=filter_test")
    assert resp.status_code == 200
    data = resp.get_json()
    for event in data["events"]:
        assert event["event_type"] == "filter_test"


def test_event_stats(client):
    client.post("/events", json={"event_type": "stat_test"})
    resp = client.get("/events/stats")
    assert resp.status_code == 200
    data = resp.get_json()
    assert "total_events" in data
    assert "by_type" in data
