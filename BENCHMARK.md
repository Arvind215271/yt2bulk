
# Yt2bulk Benchmark

## Why Yt2bulk Uses Persistent Python Workers

Yt2bulk uses persistent Python workers instead of spawning a new `yt-dlp` process for every request.

The reason is that `yt-dlp` startup cost is not free.

A normal command execution flow looks like:

```
Go
|
| spawn process
v
python
|
| initialize yt-dlp
v
fetch YouTube data
|
| process response
v
exit process
```

Every request repeats the initialization cost.

Yt2bulk changes this by keeping Python workers alive:

```
Go API Server
|
+----------------+
| Python Worker  |
| yt-dlp alive   |
+----------------+
|
+----------------+
| Python Worker  |
| yt-dlp alive   |
+----------------+
|
+----------------+
| Python Worker  |
| yt-dlp alive   |
+----------------+
```

Workers receive commands, execute `yt-dlp`, return structured data, and wait for the next request.

This removes repeated Python startup overhead and allows multiple videos to be processed concurrently.


---

# Command vs Persistent Worker Benchmark

## Single yt-dlp Command Execution

A normal `yt-dlp` command execution has additional overhead:

```
Request
|
+-- Spawn Python process
|
+-- Load yt-dlp
|
+-- Initialize extractors
|
+-- Fetch YouTube data
|
+-- Parse response
|
+-- Exit process
```

Approximate overhead:

```
Process spawning:
~1.1 seconds

Command processing / initialization:
~0.4 seconds
```

Total startup overhead:

```
~1.5 seconds/request
```

For a single request this is acceptable.

For hundreds or thousands of videos this becomes expensive because the overhead repeats.


---

# Worker Warmup

Yt2bulk workers perform warmup when starting.

Warmup may take several seconds because:

- Python worker starts
- yt-dlp modules load
- YouTube requests are made
- Internal cache files are populated

Example:

```
Worker startup:
5-10 seconds
```

However, after warmup:

```
First request:
slower

Following requests:
faster
```

The initial warmup cost is paid once instead of every request.


---

# Why Persistent Workers Are Faster

With command execution:

```
100 videos

100 x process startup
100 x yt-dlp initialization
100 x shutdown
```

With Yt2bulk:

```
Worker startup once

100 requests
100 yt-dlp executions
same workers stay alive
```

The expensive initialization is removed from every request.


---

# Benchmark Environment

Test:

```
Videos:
316

Data:
Video metadata
+
One subtitle track

Workers:
1,2,4,8,16,32,64,128
```

The benchmark:

1. Clears yt-dlp cache
2. Starts Yt2bulk
3. Warms workers
4. Fetches playlist
5. Runs export pipeline
6. Measures export time


---

# Worker Scaling Benchmark

| Workers | Export Time | Speedup |
|---------|------------:|--------:|
| 1       | 417.10s     | 1.00x |
| 2       | 180.46s     | 2.31x |
| 4       | 120.72s     | 3.45x |
| 8       | 87.85s      | 4.75x |
| 16      | 29.43s      | 14.17x |
| 32      | 18.64s      | 22.38x |
| 64      | 16.42s      | 25.40x |
| 128     | 15.63s      | 26.69x |


---

# Observations

## 1 Worker

A single worker behaves similarly to running `yt-dlp` sequentially.

Performance:

```
316 videos
~417 seconds
```

This is useful as a baseline.

---

## Increasing Workers

Adding workers significantly improves throughput:

```
1 worker:
417 seconds

16 workers:
29 seconds

128 workers:
15 seconds
```

The improvement comes from parallel execution of independent video requests.


---

# Why Not Spawn More yt-dlp Commands?

Running:

```bash
yt-dlp URL
yt-dlp URL
yt-dlp URL
...
```

creates many independent processes.

Problems:

* repeated Python startup
* repeated yt-dlp initialization
* higher memory usage
* no worker reuse
* harder resource management

Yt2bulk avoids this by managing workers internally.

---

# Warm Cache Effect

The first request is usually slower because it initializes:

* Python worker
* yt-dlp
* YouTube extraction logic
* internal cache

After warmup:

```
Average request:
~4 seconds

After cache:
~2.5 seconds for common requests
```

The warmup cost is a one-time startup cost.

---

# Conclusion

Persistent Python workers provide:

* lower request latency
* better throughput
* controlled concurrency
* reusable yt-dlp state
* reduced process spawning overhead

The worker model becomes increasingly beneficial when processing:

* playlists
* batch exports
* subtitle extraction
* transcript generation
* large YouTube datasets

Yt2bulk is designed around long-running workloads where repeatedly starting `yt-dlp` would become the bottleneck.

