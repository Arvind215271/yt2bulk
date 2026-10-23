# Deployment and Verification Plan: Incomplete Metadata Extraction

The changes to `internal/youtube/metadata/ytdlp/python/python_worker_single.py` have been implemented. 

## Changes Made
1. **Version Reporting**: Added `print(f"yt-dlp version: {yt_dlp.version.__version__}", file=sys.stderr, flush=True)` to `main()` to log the version on startup.
2. **Full Logging**: Disabled `quiet` and `no_warnings` in `YDL_OPTS`.
3. **Player Client**: Commented out the forced `android` client in `YDL_OPTS`.
4. **Cookie Support**: Added `cookiefile` configuration using `os.getenv("YTDLP_COOKIE_FILE")`.

## Next Steps for Deployment and Observation
1. **Commit and Deploy**:
   - `git add internal/youtube/metadata/ytdlp/python/python_worker_single.py`
   - `git commit -m "chore: enable debug logging, remove android client constraint, add cookie support"`
   - Push to `origin/main` to trigger the Render deployment.

2. **Monitor Logs**:
   - Access the Render dashboard for the `yt2bulk` service.
   - Observe `stderr` logs for the python worker processes.
   - Look for "yt-dlp version: ..." on startup.
   - Look for detailed warnings/errors from yt-dlp to identify why extraction is failing or incomplete.
   - If rate-limited, configure `YTDLP_COOKIE_FILE` in Render environment variables pointing to a valid cookies file.

3. **Verify Behavior**:
   - Check if metadata extraction completes successfully without the `android` constraint.
   - If still failing, analyze the full warning/error logs captured.
