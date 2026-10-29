import yt_dlp
import json

def test_different_options():
    urls = [
        "https://www.youtube.com/@YouTube",
        "https://www.youtube.com/@YouTube/videos"
    ]
    
    for url in urls:
        print(f"\n=== Testing URL: {url} ===")
        ydl_opts = {
            'extract_flat': 'in_playlist',
            'quiet': True,
            'no_warnings': True,
        }
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            try:
                info = ydl.extract_info(url, download=False)
                print("=== FULL JSON RECEIVED ===")
                if 'entries' in info:
                    print(f"Number of entries: {len(info['entries'])}")
                    for i, entry in enumerate(info['entries']):
                        print(f"Entry {i}: title={entry.get('title')}, type={entry.get('_type') or entry.get('type')}, sub_entries={len(entry.get('entries', []))}")
                        if entry.get('entries'):
                            for j, sub in enumerate(entry['entries'][:3]):
                                print(f"    Sub-entry {j}: title={sub.get('title')}, url={sub.get('url') or sub.get('webpage_url') or sub.get('original_url')}")
                else:
                    print("No entries found.")
            except Exception as e:
                print(f"Error: {e}")

if __name__ == "__main__":
    test_different_options()
