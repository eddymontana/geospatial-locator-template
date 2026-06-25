# 🌍 High-Performance Automated Geospatial Store Locator

A serverless, containerized Go microservice capable of providing sub-second nearest-neighbor geospatial searches. This architecture prioritizes execution speed, cost-efficiency, and geographical intelligence, backed by a fully automated production-grade DevOps pipeline.

🚀 **[View Live Production Demo](https://geospatial-locator-795171569538.us-central1.run.app)**

---

## 🛠️ System Architecture & CI/CD Pipeline

This project operates on a **Zero-Ops deployment philosophy**. No manual terminal intervention, cloud-shell sessions, or container pushes are required to maintain or update the service.

1. **Local Development:** Iterative updates are managed locally on Windows using VS Code.
2. **Automated Pipeline Trigger:** Running `git push origin main` securely handshakes with a GitHub Actions runner.
3. **IAM Authentication:** The runner authenticates cleanly to Google Cloud Platform using a securely masked Service Account Key (`GCP_SA_KEY`).
4. **Optimized Multi-Stage Build:** Google Cloud Build handles compilation using a multi-stage Docker execution lifecycle to keep the footprint minimal.
5. **Continuous Deployment:** The resulting image is versioned in Google Artifact Registry and rolled out instantly to an auto-scaling **Google Cloud Run** microservice.

---

## 🧰 The Tech Stack

| Layer | Technology | Key Skills Demonstrated |
| :--- | :--- | :--- |
| **Backend / Microservice** | Go (Golang) | Serverless computing, native high-concurrency routing, low startup latency. |
| **Data Kernel** | Localized GeoJSON Matrix / PostGIS Integration | Spatial data parsing, Haversine calculation vectors, mathematical scaling. |
| **Frontend Canvas** | Vanilla JS, HTML5, Custom CSS Grid | Responsive design, Google Maps JavaScript API (Satellite view, advanced markers). |
| **DevOps / Infrastructure** | Docker, GitHub Actions, GCP Artifact Registry | CI/CD pipeline orchestration, multi-stage secure container compilation. |

---

## 💡 Production Debugging & Lessons Learned

### The Case of the 1ms `500 Internal Server Error`
During deployment staging, the backend search endpoint (`/api/search`) began returning instant 1ms `500` system faults upon client requests. 

* **The Root Cause:** While local path configurations allowed the Go binary to locate data asset sheets natively on the desktop, the isolation of a standard production container structure meant the application could no longer resolve the relative `./data/` folder trees.
* **The Engineering Fix:** Restructured the final execution block of the `Dockerfile` using multi-stage artifact extraction layers. By forcing explicit filesystem directory creation (`RUN mkdir -p data static`), both front-end dependencies and GeoJSON parameters are preserved perfectly at the container layer:

```dockerfile
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

# Recreate structural directory tracking frameworks
RUN mkdir -p data static
COPY --from=builder /app/static/ ./static/
COPY --from=builder /app/data/recycling-locations.geojson ./data/

EXPOSE 8080
CMD ["./main"]


