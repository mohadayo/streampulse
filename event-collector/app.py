import os
import logging
import uuid
from datetime import datetime, timezone

from flask import Flask, request, jsonify
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("event-collector")

events_store: list[dict] = []

PROCESSOR_URL = os.getenv("PROCESSOR_URL", "http://event-processor:8081")


@app.route("/health", methods=["GET"])
def health():
    now = datetime.now(timezone.utc).isoformat()
    return jsonify({"status": "healthy", "service": "event-collector", "timestamp": now})


@app.route("/events", methods=["POST"])
def collect_event():
    data = request.get_json()
    if not data:
        logger.warning("Received empty event payload")
        return jsonify({"error": "Request body must be valid JSON"}), 400

    if "event_type" not in data:
        logger.warning("Missing event_type in payload")
        return jsonify({"error": "event_type is required"}), 400

    event = {
        "id": str(uuid.uuid4()),
        "event_type": data["event_type"],
        "payload": data.get("payload", {}),
        "source": data.get("source", "unknown"),
        "timestamp": datetime.now(timezone.utc).isoformat(),
    }

    events_store.append(event)
    logger.info("Collected event %s of type %s", event["id"], event["event_type"])

    return jsonify({"message": "Event collected", "event_id": event["id"]}), 201


@app.route("/events", methods=["GET"])
def list_events():
    event_type = request.args.get("event_type")
    limit = request.args.get("limit", 100, type=int)

    filtered = events_store
    if event_type:
        filtered = [e for e in events_store if e["event_type"] == event_type]

    return jsonify({"events": filtered[-limit:], "total": len(filtered)})


@app.route("/events/stats", methods=["GET"])
def event_stats():
    type_counts: dict[str, int] = {}
    for event in events_store:
        t = event["event_type"]
        type_counts[t] = type_counts.get(t, 0) + 1

    return jsonify({
        "total_events": len(events_store),
        "by_type": type_counts,
    })


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    debug = os.getenv("FLASK_DEBUG", "false").lower() == "true"
    logger.info("Starting event-collector on port %d", port)
    app.run(host="0.0.0.0", port=port, debug=debug)
