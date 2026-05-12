import request from "supertest";
import { app } from "./index";

describe("API Gateway", () => {
  describe("GET /health", () => {
    it("returns healthy status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.status).toBe("healthy");
      expect(res.body.service).toBe("api-gateway");
      expect(res.body.timestamp).toBeDefined();
      expect(res.body.upstreams).toBeDefined();
    });
  });

  describe("POST /api/v1/events", () => {
    it("returns 502 when collector is unreachable", async () => {
      const res = await request(app)
        .post("/api/v1/events")
        .send({ event_type: "test", payload: {} });
      expect(res.status).toBe(502);
      expect(res.body.error).toBeDefined();
    });
  });

  describe("GET /api/v1/events", () => {
    it("returns 502 when collector is unreachable", async () => {
      const res = await request(app).get("/api/v1/events");
      expect(res.status).toBe(502);
    });
  });

  describe("GET /api/v1/events/stats", () => {
    it("returns 502 when collector is unreachable", async () => {
      const res = await request(app).get("/api/v1/events/stats");
      expect(res.status).toBe(502);
    });
  });

  describe("GET /api/v1/processed", () => {
    it("returns 502 when processor is unreachable", async () => {
      const res = await request(app).get("/api/v1/processed");
      expect(res.status).toBe(502);
    });
  });

  describe("GET /unknown", () => {
    it("returns 404 for unknown routes", async () => {
      const res = await request(app).get("/unknown");
      expect(res.status).toBe(404);
      expect(res.body.error).toBe("Not found");
    });
  });
});
