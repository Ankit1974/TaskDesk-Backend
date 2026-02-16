# Docker Setup Guide — TaskDesk Backend

## Prerequisites

- **Docker** installed on your machine
  - Mac: [Download Docker Desktop](https://www.docker.com/products/docker-desktop/)
  - Linux: `sudo apt install docker.io` (Ubuntu/Debian)
  - Verify: `docker --version`

---

## Project Structure (Docker-related files)

```
TaskDesk-Backend/
├── Dockerfile          # Multi-stage build (build + run)
├── .dockerignore       # Files excluded from Docker image
├── .env                # Your environment variables (NOT copied into image)
└── .env.example        # Template for required env vars
```

---

## Step 1: Set Up Environment Variables

Make sure your `.env` file exists with all required values:

```env
APP_PORT=8080
ENV=production

SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_JWT_SECRET=your_jwt_secret
DATABASE_URL=postgresql://postgres:your_password@db.your-project.supabase.co:5432/postgres
```

---

## Step 2: Build the Docker Image

Run this from the project root:

```bash
docker build -t taskdesk-api .
```

**What this does:**
1. **Build stage** — Uses `golang:1.25-alpine` to compile your Go code into a binary
2. **Run stage** — Copies just the binary into a minimal `alpine` image (~15MB)

---

## Step 3: Run the Container

```bash
docker run -d \
  --name taskdesk-api \
  --env-file .env \
  -p 8080:8080 \
  taskdesk-api
```

**Flags explained:**
| Flag | Purpose |
|------|---------|
| `-d` | Run in background (detached mode) |
| `--name taskdesk-api` | Name the container for easy reference |
| `--env-file .env` | Load environment variables from your `.env` file |
| `-p 8080:8080` | Map port 8080 on your machine to port 8080 in the container |

Your API is now running at: **http://localhost:8080**

---

## Step 4: Verify It's Working

```bash
# Check container is running
docker ps

# Check health endpoint
curl http://localhost:8080/api/v1/health

# View logs
docker logs taskdesk-api

# View logs in real-time
docker logs -f taskdesk-api
```

---

## Common Docker Commands

```bash
# Stop the container
docker stop taskdesk-api

# Start it again
docker start taskdesk-api

# Restart
docker restart taskdesk-api

# Remove the container (must stop first)
docker stop taskdesk-api && docker rm taskdesk-api

# Rebuild after code changes
docker build -t taskdesk-api . && \
docker stop taskdesk-api && \
docker rm taskdesk-api && \
docker run -d --name taskdesk-api --env-file .env -p 8080:8080 taskdesk-api
```

---

## Deploying to Render with Docker

Once Docker is set up, deploying to Render is straightforward:

1. Push your code (with `Dockerfile`) to GitHub
2. Go to [render.com](https://render.com) → **New Web Service**
3. Connect your GitHub repo
4. Render auto-detects the `Dockerfile`
5. Add your environment variables in Render's dashboard:
   - `APP_PORT` = `8080`
   - `ENV` = `production`
   - `DATABASE_URL` = your Supabase connection string
   - `SUPABASE_URL` = your Supabase project URL
   - `SUPABASE_ANON_KEY` = your anon key
   - `SUPABASE_JWT_SECRET` = your JWT secret
6. Deploy

Render will build your Docker image and give you a public URL like:
`https://taskdesk-api.onrender.com`

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `port is already allocated` | Another process is using port 8080. Stop it or use `-p 3000:8080` to map to a different port |
| Container exits immediately | Check logs: `docker logs taskdesk-api` — likely a missing env variable |
| Can't connect to Supabase DB | Make sure `DATABASE_URL` in `.env` is correct and Supabase project is active |
| `image not found` | Run `docker build -t taskdesk-api .` first |
| Changes not reflected | You need to rebuild: `docker build -t taskdesk-api .` |
