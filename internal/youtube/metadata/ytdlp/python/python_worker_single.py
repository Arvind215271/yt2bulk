#!/usr/bin/env python3

import gc
import os
import struct
import sys
import time

import orjson
import yt_dlp
import traceback


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
    "cachedir": True,
    "check_formats": False,
    "ignore_no_formats_error": True,
    "socket_timeout": 10,
    "extractor_retries": 0,
    "retries": 0,
    "lazy_playlist": True,
    "noprogress": True,
    # "cookiefile": "/home/arvind-saini/Projects/askito/cookies.txt",    
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


def extract_with_retry(ydl, url, download=False, max_retries=3, initial_delay=2.0):
    delay = initial_delay
    for attempt in range(max_retries):
        try:
            return ydl.extract_info(url, download=download)
        except Exception as e:
            err_msg = str(e)
            is_rate_limit = "429" in err_msg or "Too Many Requests" in err_msg or "HTTP Error" in err_msg
            if attempt == max_retries - 1 or not is_rate_limit:
                raise
            time.sleep(delay)
            delay *= 2.0


def download_with_retry(ydl, url_list, max_retries=3, initial_delay=2.0):
    delay = initial_delay
    for attempt in range(max_retries):
        try:
            ydl.download(url_list)
            return
        except Exception as e:
            err_msg = str(e)
            is_rate_limit = "429" in err_msg or "Too Many Requests" in err_msg or "HTTP Error" in err_msg
            if attempt == max_retries - 1 or not is_rate_limit:
                raise
            time.sleep(delay)
            delay *= 2.0


def main():
    ydl = yt_dlp.YoutubeDL(YDL_OPTS)

    gc.collect()
    gc.freeze()

    requests = 0

    while True:

        try:
            req = recv()
        except EOFError:
            break
        except Exception:
            traceback.print_exc(file=sys.stderr)
            break

        try:

            cmd = req.get("cmd")

            if cmd == "playlist":

                playlist_id = req["playlist_id"]

                playlist_url = (
                    "https://www.youtube.com/playlist?list="
                    + playlist_id
                )

                old_extract_flat = ydl.params.get("extract_flat")

                try:
                    ydl.params["extract_flat"] = True

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

                send({
                    "ok": True,
                    "data": info,
                })

                del info

                continue

            if cmd == "playlist_metadata":

                playlist_id = req["playlist_id"]

                playlist_url = (
                    "https://www.youtube.com/playlist?list="
                    + playlist_id
                )

                # Set extract_flat to True to avoid enumerating/downloading playlist entries
                # and ensure we only get playlist-level metadata.
                old_extract_flat = ydl.params.get("extract_flat")
                old_playlist_items = ydl.params.get("playlist_items")

                try:
                    ydl.params["extract_flat"] = True
                    # Setting playlist_items to a single entry that doesn't exist
                    # or an empty range might be an alternative,
                    # but extract_flat=True is typically sufficient for metadata.
                    # We will ensure no items are fetched.
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

            if cmd == "playlist_surface_metadata":

                playlist_id = req["playlist_id"]

                playlist_url = (
                    "https://www.youtube.com/playlist?list="
                    + playlist_id
                )

                old_extract_flat = ydl.params.get("extract_flat")

                try:
                    ydl.params["extract_flat"] = True

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

                try:
                    ydl.params["writesubtitles"] = subtitle_type == "manual"
                    ydl.params["writeautomaticsub"] = subtitle_type == "automatic"
                    ydl.params["subtitleslangs"] = [language]
                    ydl.params["subtitlesformat"] = fmt
                    ydl.params["outtmpl"] = outtmpl

                    download_with_retry(ydl, [YT_URL + video_id])

                finally:
                    ydl.params["outtmpl"] = old_outtmpl

                send({
                    "ok": True,
                })

                continue

            info = extract_with_retry(
                ydl,
                YT_URL + req["video_id"],
                download=False,
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

            del req

            requests += 1

            if requests >= 50:
                gc.collect()
                requests = 0


if __name__ == "__main__":
    main()
