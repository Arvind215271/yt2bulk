# Yt2bulk

Yt2bulk is a Go application for extracting and processing YouTube data.

It can fetch video and playlist metadata, download subtitles, generate transcripts, analyze transcript statistics, and export the results in different formats.

The application uses Go for the HTTP API and request handling. It runs persistent Python workers that use `yt-dlp` to fetch data from YouTube. Keeping the workers alive avoids starting a new Python process for every request, which reduces overhead when processing many videos.

Check out our [Benchmarks](BENCHMARK.md) and [Player Client Compatibility](PLAYER_CLIENT.md) for more details on performance and configuration.

# Setup & Running Instructions

## 1. Backend (Go Server) Setup & Running

### Requirements
Before running Yt2bulk, install:
- Go 1.20 or later
- Python 3.8 or later
- pip
- python3-venv
- FFmpeg (recommended by `yt-dlp`)

### Installation

#### Clone the repository
```bash
git clone https://github.com/your-username/yt2bulk.git
cd yt2bulk
```

#### Install Go dependencies
```bash
go mod download
```

#### Create a Python virtual environment
```bash
python3 -m venv venv

# Linux/macOS
source venv/bin/activate

# Windows
venv\Scripts\activate
```

#### Install Python dependencies
```bash
pip install --upgrade pip
pip install yt-dlp orjson
```
Yt2bulk starts Python workers when the server launches. These workers require `yt-dlp` and `orjson`, so make sure they are installed inside the virtual environment.

### Configuration ([`.env`](.env))
Copy the sample environment configuration file:
```bash
cp .env.sample .env
```

Open [`.env`](.env) and customize your configuration parameters:
```env
APP_ENV=development
PORT=8080

# Optional: YouTube Data API v3 key if using the YouTube Data API provider
# YOUTUBE_API_KEY=your_api_key_here

# Cache directory configuration
YTDLP_CACHE_DIR=./.cache/ytdlp
YTDLP_CACHE_TTL_DAYS=28
YTDLP_CACHE_MAX_FILES=2000

# Number of concurrent Python workers for yt-dlp processing
# Each worker consumes ~50-60 MB RAM. Adjust based on your system capacity.
PYTHON_WORKERS=16
```

### Running Backend
Start the server:
```bash
go run main.go
```

Or build a binary:
```bash
go build -o yt2bulk main.go
./yt2bulk
```

## 2. Frontend (Next.js App) Setup & Running

The frontend is a Next.js application located in the [`frontend/`](frontend/) directory. It interacts with the Go backend APIs listed below to visualize the data and orchestrate requests.

### Prerequisites
- Node.js (v18 or later)
- npm

### Setup
```bash
cd frontend
npm install
```

### Running in Development
```bash
cd frontend
npm run dev
```
The application will be available at `http://localhost:3000`.

### Building for Production
```bash
cd frontend
npm run build
npm run start
```

# Performance & Architecture

Yt2bulk uses **persistent Python workers** instead of spawning a new `yt-dlp` process for every request. This eliminates repeated Python startup and initialization overhead (~1.5s per request), enabling high-throughput parallel video processing.

## Benchmark Summary (316 Videos Playlist Export)

| Workers | Export Time | Speedup |
|---------|------------:|--------:|
| 1       | 417.10s     | 1.00x   |
| 16      | 29.43s      | 14.17x  |
| 128     | 15.63s      | 26.69x  |

See [BENCHMARK.md](BENCHMARK.md) for full architectural details and scaling metrics.

# API Endpoints, Parameters & Models Reference

Once the server is running at `http://localhost:8080`, you can access the following REST endpoints exposed by [`internal/api`](internal/api/):

### 1. Export API (`/export`)

#### Export Resources
- **Endpoint**: `POST /export`
- **Request Body Model**: [`internal/api/export/models.go`](internal/api/export/models.go)
  ```json
  {
    "inputs": ["string"],
    "subtitle": {
      "video_id": "string",
      "language": "string",
      "type": "string",
      "format": "string"
    },
    "preferences": [
      {
        "language": "string",
        "type": "string" // options: "manual", "automatic", "manual>automatic", "automatic>manual"
      }
    ],
    "transcript": {
      "language": "string",
      "type": "string"
    },
    "signal": {
      "analysis": "string", // e.g. "word-stats", "window-stats"
      "language": "string",
      "type": "string",
      "use_heavy_stopwords": true,
      "min_freq": 0
    },
    "format": "string", // options: "json", "csv", "markdown", "excel", "yaml", "xml"
    "fields": ["string"]
  }
  ```
- **Available Fields**:
  - Metadata: `id`, `errors`, `title`, `description`, `channel_id`, `channel_title`, `thumbnails`, `published_at`, `duration`, `duration_seconds`, `duration_minutes`, `duration_timestamp`, `view_count`, `like_count`, `comment_count`, `tags`, `category_id`, `caption_available`, `privacy_status`, `live_broadcast_status`
  - Description: `description_chapters`, `description_links`, `description_emails`, `description_cleaned`
  - Transcript/Subtitle: `transcript_text`, `transcript_signal`, `subtitle_metadata`

- **Example Request**:
  ```bash
  curl -X POST "http://localhost:8080/export" \
    -H "Content-Type: application/json" \
    -d '{
      "inputs": ["https://www.youtube.com/watch?v=dQw4w9WgXcQ"],
      "format": "json",
      "fields": ["id", "title", "duration"]
    }'
  ```

### 2. Playlist API (`/playlist`)

#### Expand Playlist Videos
- **Endpoint**: `POST /playlist/videos`
- **Request Body Model**: [`internal/api/playlist/models.go`](internal/api/playlist/models.go)
  ```json
  {
    "url": "string",
    "provider": "string", // options: "ytdlp", "youtube_api"
    "output": "string"    // options: "id", "url", "both"
  }
  ```
- **Response Model**:
  ```json
  {
    "playlist_id": "string",
    "total": 0,
    "videos": [
      {
        "id": "string",
        "url": "string",
        "position": 0,
        "added_at": "string"
      }
    ]
  }
  ```
- **Example Request**:
  ```bash
  curl -X POST "http://localhost:8080/playlist/videos" \
    -H "Content-Type: application/json" \
    -d '{
      "url": "https://www.youtube.com/playlist?list=PL_EXAMPLE",
      "output": "both"
    }'
  ```

### 3. Subtitle API (`/subtitle`)

#### Get Subtitle Options
- **Endpoint**: `POST /subtitle/options`
- **Request Body Model**: [`internal/api/subtitle/models.go`](internal/api/subtitle/models.go)
  ```json
  {
    "inputs": ["string"]
  }
  ```
- **Example Request**:
  ```bash
  curl -X POST "http://localhost:8080/subtitle/options" \
    -H "Content-Type: application/json" \
    -d '{
      "inputs": ["https://www.youtube.com/watch?v=dQw4w9WgXcQ"]
    }'
  ```

#### Download Subtitle
- **Endpoint**: `POST /subtitle/download`
- **Request Body Model**: [`internal/api/subtitle/models.go`](internal/api/subtitle/models.go)
  ```json
  {
    "inputs": ["string"],
    "preferences": [
      {
        "language": "string",
        "type": "string" // options: "manual", "automatic", "manual>automatic", "automatic>manual"
      }
    ],
    "format": "string" // options: "json3", "vtt", "srt", "srv1", "srv2", "srv3", "ttml"
  }
  ```
- **Example Request**:
  ```bash
  curl -X POST "http://localhost:8080/subtitle/download" \
    -H "Content-Type: application/json" \
    -d '{
      "inputs": ["https://www.youtube.com/watch?v=dQw4w9WgXcQ"],
      "format": "vtt"
    }'
  ```

### 4. Transcript API (`/transcript`)

#### Get Transcript
- **Endpoint**: `POST /transcript`
- **Request Body Model**: [`internal/api/transcript/models.go`](internal/api/transcript/models.go)
  ```json
  {
    "inputs": ["string"],
    "preferences": [
      {
        "language": "string",
        "type": "string" // options: "manual", "automatic", "manual>automatic", "automatic>manual"
      }
    ]
  }
  ```
- **Example Request**:
  ```bash
  curl -X POST "http://localhost:8080/transcript" \
    -H "Content-Type: application/json" \
    -d '{
      "inputs": ["https://www.youtube.com/watch?v=dQw4w9WgXcQ"]
    }'
  ```
