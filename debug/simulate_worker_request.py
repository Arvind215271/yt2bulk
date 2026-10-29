import sys
import struct
import json
import subprocess
import os

# Use python from venv explicitly for subprocess as well
python_path = os.path.abspath("./venv/bin/python3")

PACK = struct.Struct(">I")

def send_request(process, obj):
    payload = json.dumps(obj).encode("utf-8")
    process.stdin.write(PACK.pack(len(payload)))
    process.stdin.write(payload)
    process.stdin.flush()

def read_response(process):
    hdr = bytearray(4)
    if process.stdout.readinto(hdr) != 4:
        return None
    size = struct.unpack(">I", hdr)[0]
    buf = bytearray(size)
    
    bytes_read = 0
    while bytes_read < size:
        chunk = process.stdout.read(size - bytes_read)
        if not chunk:
            break
        buf[bytes_read:bytes_read+len(chunk)] = chunk
        bytes_read += len(chunk)
        
    return json.loads(buf.decode("utf-8"))

worker_path = "internal/youtube/metadata/ytdlp/python/python_worker_single.py"
process = subprocess.Popen(
    [python_path, worker_path],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=sys.stderr,
    bufsize=0
)

request = {
    "cmd": "channel",
    "channel_url": "https://www.youtube.com/@krissk3b"
}

print(f"Sending request: {request}")
send_request(process, request)

response = read_response(process)
with open("debug/channel_worker_response.json", "w") as f:
    json.dump(response, f, indent=2, default=str)
print(f"Response saved to debug/channel_worker_response.json")

process.stdin.close()
process.stdout.close()
process.terminate()
