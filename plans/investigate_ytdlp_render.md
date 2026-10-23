# Investigation Plan: Incomplete Metadata Extraction on Render

## Objectives
1. **Report yt-dlp version**: Log the active `yt-dlp` version on worker initialization so Render logs explicitly display what version is running in production.
2. **Enable full extraction warnings**: Modify `YDL_OPTS` in [`python_worker_single.py`](internal/youtube/metadata/ytdlp/python/python_worker_single.py) with `"quiet": False` and `"no_warnings": False` to capture detailed yt-dlp diagnostics in stderr.
3. **Test default player-client selection**: Remove the forced `["android"]` player client in `extractor_args` to allow `yt-dlp` to fallback or negotiate clients naturally.
4. **Prepare for secure cookie testing**: Support configuring `cookiefile` via environment variable or path configuration to test authenticated requests against YouTube rate limits/bot detection.

---

## Detailed Steps & Changes

### 1. Version Reporting (`python_worker_single.py`)
- Import `yt_dlp`.
- In `main()`, upon startup, print:
  ```python
  print(f"yt-dlp version: {yt_dlp.version.__version__}", file=sys.stderr, flush=True)
  ```
- Optionally handle a `cmd == "version"` request if needed, but startup logging ensures it's captured immediately when the worker spawns.

### 2. Full Extraction Warnings & Logging
- In `YDL_OPTS`:
  ```python
  "quiet": False,
  "no_warnings": False,
  ```
- This ensures all warnings, extraction steps, and player client fallback attempts are logged to `stderr`, which the Go `PythonWorker` reads and routes to the structured logger (`workerLogger.Warn`).

### 3. Default Player-Client Selection
- Remove or comment out:
  ```python
  # "extractor_args": {
  #     "youtube": {
  #         "player_client": ["android"],
  #     }
  # },
  ```
- This allows `yt-dlp` to use its default client negotiation strategy (web, mweb, ios, tv, android, etc.).

### 4. Secure Cookie Testing Preparation
- Allow loading cookies from an environment variable path (`YTDLP_COOKIE_FILE` or similar), or default to a secure path if present:
  ```python
  cookie_file = os.getenv("YTDLP_COOKIE_FILE")
  if cookie_file and os.path.exists(cookie_file):
      YDL_OPTS["cookiefile"] = cookie_file
  ```

---

## Verification & Deployment Plan
1. Apply code modifications to [`internal/youtube/metadata/ytdlp/python/python_worker_single.py`](internal/youtube/metadata/ytdlp/python/python_worker_single.py).
2. Run local tests / verify worker startup.
3. Commit and push changes to trigger a Render deployment.
4. Monitor Render logs for yt-dlp version, warnings, and extraction behavior.
