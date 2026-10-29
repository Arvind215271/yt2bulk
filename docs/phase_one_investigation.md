# Phase One: `yt-dlp` Channel Extraction Investigation

## 1. Goal
Verify `yt-dlp` behavior for channel URL extraction to determine if we can perform lightweight video discovery and if pagination/limits need special handling.

## 2. Methodology
- Used a Python script utilizing `yt_dlp` with `extract_flat=True`.
- Tested against `https://www.youtube.com/@YouTube` (both the base URL and the `/videos` tab).

## 3. Findings

### Metadata Availability
- **Base URL (@YouTube):** Returns channel-level metadata (title, ID) and a few "collection" entries (e.g., "Videos", "Live", "Shorts"). It does *not* immediately list all videos.
- **Videos Tab URL (@YouTube/videos):** Returns a flat listing of videos from the channel.

### Video Listing
- **Pagination:** `yt-dlp` handled the continuation-based pagination automatically.
- **Completeness:** The initial call to the `/videos` tab returned a substantial list of entries (198 in the test). It appears to retrieve a significant portion, possibly all, of the recent videos depending on channel size.
- **Metadata:** The `extract_flat` mode returns basic information: `title`, `id`, and `url` for each video. This is sufficient for lightweight discovery.

### Performance
- The retrieval was fast for a channel with hundreds of videos.

### Conclusion for Implementation
- We should point the backend channel discovery to the channel's `/videos` URL to ensure we get the video listing directly.
- The returned list is suitable for the "lightweight listing" requirement.
- No complex custom pagination logic seems required initially; `yt-dlp` handles the bulk of it.
