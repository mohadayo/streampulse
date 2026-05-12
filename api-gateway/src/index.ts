import express, { Request, Response, NextFunction } from "express";
import cors from "cors";
import axios, { AxiosError } from "axios";

const app = express();
app.use(cors());
app.use(express.json());

const COLLECTOR_URL = process.env.COLLECTOR_URL || "http://event-collector:8080";
const PROCESSOR_URL = process.env.PROCESSOR_URL || "http://event-processor:8081";
const PORT = parseInt(process.env.PORT || "8082", 10);
const LOG_LEVEL = process.env.LOG_LEVEL || "info";

function log(level: string, message: string, meta?: Record<string, unknown>): void {
  const levels = ["debug", "info", "warn", "error"];
  if (levels.indexOf(level) < levels.indexOf(LOG_LEVEL)) return;
  const entry = {
    timestamp: new Date().toISOString(),
    level,
    service: "api-gateway",
    message,
    ...meta,
  };
  console.log(JSON.stringify(entry));
}

app.get("/health", (_req: Request, res: Response) => {
  res.json({
    status: "healthy",
    service: "api-gateway",
    timestamp: new Date().toISOString(),
    upstreams: {
      collector: COLLECTOR_URL,
      processor: PROCESSOR_URL,
    },
  });
});

app.post("/api/v1/events", async (req: Request, res: Response) => {
  try {
    log("info", "Forwarding event to collector", { event_type: req.body?.event_type });

    const collectorResp = await axios.post(`${COLLECTOR_URL}/events`, req.body, {
      timeout: 5000,
    });

    const eventId = collectorResp.data.event_id;

    try {
      await axios.post(
        `${PROCESSOR_URL}/process`,
        { ...req.body, id: eventId },
        { timeout: 5000 }
      );
      log("info", "Event processed successfully", { event_id: eventId });
    } catch (procErr) {
      log("warn", "Processor unavailable, event collected but not processed", { event_id: eventId });
    }

    res.status(201).json({
      message: "Event ingested",
      event_id: eventId,
    });
  } catch (err) {
    const axErr = err as AxiosError;
    if (axErr.response) {
      log("error", "Collector returned error", { status: axErr.response.status });
      res.status(axErr.response.status).json(axErr.response.data);
    } else {
      log("error", "Collector unreachable", { error: axErr.message });
      res.status(502).json({ error: "Upstream service unavailable" });
    }
  }
});

app.get("/api/v1/events", async (_req: Request, res: Response) => {
  try {
    const resp = await axios.get(`${COLLECTOR_URL}/events`, {
      params: _req.query,
      timeout: 5000,
    });
    res.json(resp.data);
  } catch {
    log("error", "Failed to fetch events from collector");
    res.status(502).json({ error: "Upstream service unavailable" });
  }
});

app.get("/api/v1/events/stats", async (_req: Request, res: Response) => {
  try {
    const resp = await axios.get(`${COLLECTOR_URL}/events/stats`, { timeout: 5000 });
    res.json(resp.data);
  } catch {
    log("error", "Failed to fetch stats from collector");
    res.status(502).json({ error: "Upstream service unavailable" });
  }
});

app.get("/api/v1/processed", async (_req: Request, res: Response) => {
  try {
    const resp = await axios.get(`${PROCESSOR_URL}/processed`, { timeout: 5000 });
    res.json(resp.data);
  } catch {
    log("error", "Failed to fetch processed events");
    res.status(502).json({ error: "Upstream service unavailable" });
  }
});

app.use((_req: Request, res: Response) => {
  res.status(404).json({ error: "Not found" });
});

app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  log("error", "Unhandled error", { error: err.message });
  res.status(500).json({ error: "Internal server error" });
});

export { app };

if (require.main === module) {
  app.listen(PORT, "0.0.0.0", () => {
    log("info", `API Gateway started on port ${PORT}`);
  });
}
