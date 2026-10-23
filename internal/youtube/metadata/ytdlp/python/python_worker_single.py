#!/usr/bin/env python3

import ctypes
import gc
import os
import random
import struct
import sys
import time
import traceback

import orjson
from yt_dlp.utils import DownloadError
import yt_dlp

# --- MEMORY MANAGEMENT GUARD ---
# Load C standard library to allow forcing glibc to return memory to the OS
try:
  libc = ctypes.CDLL("libc.so.6")
except Exception:
  libc = None


def release_memory():
  """Forces Python GC collection and instructs glibc to release

  unused heap memory back to the Linux kernel immediately.
  """
  gc.collect()
  if libc:
    libc.malloc_trim(0)


# -------------------------------

stdin = sys.stdin.buffer
stdout = sys.stdout.buffer

PACK = struct.Struct(">I")
pack = PACK.pack
unpack = PACK.unpack

YT_URL = "https://www.youtube.com/watch?v="
WARMUP_URL = YT_URL + "dQw4w9WgXcQ"

YDL_OPTS = {
    "skip_download": True,
    "quiet": True,
    "no_warnings": True,
    "ignoreconfig": True,
    "cachedir": None,
    "check_formats": False,
    "ignore_no_formats_error": True,
    "socket_timeout": 10,
    "extractor_retries": 0,
    "retries": 0,
    "lazy_playlist": True,
    "noprogress": True,
    # "cookiefile": "/home/arvind-saini/Projects/yt2bulk/cookies.txt",
    "extractor_args": {
        "youtube": {
            "player_client": ["android"],
        }
    },
}


def recv():
  hdr = bytearray(4)
  if stdin.readinto(hdr) != 4:
    raise EOFError

  size = unpack(hdr)[0]

  buf = bytearray(size)
  view = memoryview(buf)

  while view:
    n = stdin.readinto(view)
    if not n:
      raise EOFError
    view = view[n:]

  return orjson.loads(buf)


def send(obj):
  payload = orjson.dumps(obj)
  stdout.write(pack(len(payload)))
  stdout.write(payload)
  stdout.flush()


def extract_with_retry(
    ydl, url, download=False, max_retries=2, initial_delay=0.5, max_delay=2.0
):
  for attempt in range(max_retries):
    try:
      return ydl.extract_info(url, download=download)
    except Exception as e:
      error = str(e)
      is_rate_limit = "429" in error or "Too Many Requests" in error

      if not is_rate_limit or attempt == max_retries - 1:
        raise

      delay = min(initial_delay * (2**attempt), max_delay)
      delay *= random.uniform(0.8, 1.2)

      print(
          f"Rate limited by YouTube. Retry {attempt + 1}/{max_retries} in"
          f" {delay:.1f}s",
          file=sys.stderr,
          flush=True,
      )
      time.sleep(delay)


def download_with_retry(
    ydl, url_list, max_retries=2, initial_delay=0.5, max_delay=2.0
):
  for attempt in range(max_retries):
    try:
      ydl.download(url_list)
      return
    except DownloadError as e:
      error = str(e)
      is_rate_limit = "429" in error or "Too Many Requests" in error

      if not is_rate_limit or attempt == max_retries - 1:
        raise

      delay = min(initial_delay * (2**attempt), max_delay)
      delay *= random.uniform(0.8, 1.2)

      print(
          f"Rate limited by YouTube. Retry {attempt + 1}/{max_retries} in"
          f" {delay:.1f}s",
          file=sys.stderr,
          flush=True,
      )
      time.sleep(delay)


def main():
  ydl = yt_dlp.YoutubeDL(YDL_OPTS)

  # Let Python manage standard GC thresholds naturally;
  # we rely on explicit trimming via release_memory() instead.
  release_memory()

  requests = 0

  while True:
    try:
      req = recv()
    except EOFError:
      break
    except Exception:
      traceback.print_exc(file=sys.stderr)
      break

    # Track if the current request was a heavy playlist command
    is_heavy_cmd = False

    try:
      cmd = req.get("cmd")

      if cmd in ["playlist", "playlist_metadata", "playlist_surface_metadata"]:
        is_heavy_cmd = True
        playlist_id = req["playlist_id"]
        playlist_url = "https://www.youtube.com/playlist?list=" + playlist_id

        old_extract_flat = ydl.params.get("extract_flat")
        old_playlist_items = ydl.params.get("playlist_items")

        try:
          ydl.params["extract_flat"] = True
          if cmd == "playlist_metadata":
            ydl.params["playlist_items"] = "0"

          info = extract_with_retry(
              ydl,
              playlist_url,
              download=False,
          )

        finally:
          if old_extract_flat is None:
            ydl.params.pop("extract_flat", None)
          else:
            ydl.params["extract_flat"] = old_extract_flat

          if old_playlist_items is None:
            ydl.params.pop("playlist_items", None)
          else:
            ydl.params["playlist_items"] = old_playlist_items

        send({
            "ok": True,
            "data": info,
        })

        del info
        continue

      if cmd == "warmup":
        extract_with_retry(ydl, WARMUP_URL, download=False)
        send({"ok": True})
        continue

      if cmd == "subtitle":
        video_id = req["video_id"]
        language = req.get("language", "en")
        subtitle_type = req.get("type", "manual")
        fmt = req.get("format", "json3")
        output_path = req.get("output_path")

        if not output_path:
          raise RuntimeError("output_path missing in subtitle request")

        output_dir = os.path.dirname(output_path)
        os.makedirs(output_dir, exist_ok=True)

        old_outtmpl = dict(ydl.params["outtmpl"])
        outtmpl = dict(old_outtmpl)
        outtmpl["subtitle"] = output_path

        old_writesubtitles = ydl.params.get("writesubtitles")
        old_writeautomaticsub = ydl.params.get("writeautomaticsub")
        old_subtitleslangs = ydl.params.get("subtitleslangs")
        old_subtitlesformat = ydl.params.get("subtitlesformat")

        try:
          ydl.params["writesubtitles"] = subtitle_type == "manual"
          ydl.params["writeautomaticsub"] = subtitle_type == "automatic"
          ydl.params["subtitleslangs"] = [language]
          ydl.params["subtitlesformat"] = fmt
          ydl.params["outtmpl"] = outtmpl

          download_with_retry(ydl, [YT_URL + video_id])

        finally:
          ydl.params["outtmpl"] = old_outtmpl
          if old_writesubtitles is None:
            ydl.params.pop("writesubtitles", None)
          else:
            ydl.params["writesubtitles"] = old_writesubtitles

          if old_writeautomaticsub is None:
            ydl.params.pop("writeautomaticsub", None)
          else:
            ydl.params["writeautomaticsub"] = old_writeautomaticsub

          if old_subtitleslangs is None:
            ydl.params.pop("subtitleslangs", None)
          else:
            ydl.params["subtitleslangs"] = old_subtitleslangs

          if old_subtitlesformat is None:
            ydl.params.pop("subtitlesformat", None)
          else:
            ydl.params["subtitlesformat"] = old_subtitlesformat

        send({"ok": True})
        continue

      info = extract_with_retry(
          ydl,
          YT_URL + req["video_id"],
          download=False,
      )

      print(
          {
              "video_id": req["video_id"],
              "duration": info.get("duration"),
              "manual_subtitle_languages": list(
                  (info.get("subtitles") or {}).keys()
              ),
              "automatic_subtitle_languages": list(
                  (info.get("automatic_captions") or {}).keys()
              ),
          },
          file=sys.stderr,
          flush=True,
      )

      send({
          "ok": True,
          "data": info,
      })

      del info

    except Exception:
      send({
          "ok": False,
          "error": traceback.format_exc(),
      })

    finally:
      if "req" in locals():
        del req

      requests += 1

      # Forcefully release memory pages back to the OS after every single turn
      release_memory()

      # If a heavy playlist was processed, or every 10 general requests,
      # completely rebuild the YoutubeDL instance to clear internal extractor caches.
      if is_heavy_cmd or requests >= 10:
        del ydl
        release_memory()
        ydl = yt_dlp.YoutubeDL(YDL_OPTS)
        requests = 0


if __name__ == "__main__":
  main()
