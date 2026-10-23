FROM golang:1.25.8-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/yt2bulk .


FROM python:3.12-slim-bookworm

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg nodejs npm \
    && node --version \
    && npm --version \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN python -m venv /app/venv \
    && /app/venv/bin/pip install --no-cache-dir -r requirements.txt

COPY --from=builder /out/yt2bulk /app/yt2bulk

COPY internal/youtube/metadata/ytdlp/python/python_worker_single.py \
     /app/internal/youtube/metadata/ytdlp/python/python_worker_single.py

CMD ["/app/yt2bulk"]
