# Preflight API Migration Discovery & Architecture Plan

## Overview
This plan outlines the discovery and architecture for migrating the preflight API handler ([[`internal/api/preflight/handler.go`](internal/api/preflight/handler.go:1)]) to use the new preflight service ([[`internal/youtube/preflight/service.go`](internal/youtube/preflight/service.go:1)]).

## Discovery Findings

1. **Existing Handler (`[`internal/api/preflight/handler.go`](internal/api/preflight/handler.go:1)`)**:
   - Binds [`PreflightRequest`](internal/api/preflight/models.go:3) from the HTTP request body.
   - Validates that `Inputs` are provided and parses each input string using [`youtubeurl.Parse`](internal/youtube/input/parser.go:1).
   - Constructs a [`preflight.Request`](internal/youtube/preflight/model.go:12) containing `Inputs []*youtubeurl.YouTubeInput`.
   - Calls [`h.preflightService.Check(ctx, preflightReq)`](internal/api/preflight/handler.go:45).
   - Maps the resulting [`preflight.Result`](internal/youtube/preflight/model.go:16) into [`PreflightResponse`](internal/api/preflight/models.go:14).

2. **Existing Models (`[`internal/api/preflight/models.go`](internal/api/preflight/models.go:1)`)**:
   - [`PreflightRequest`](internal/api/preflight/models.go:3): contains `Inputs []string`.
   - [`PreflightStatsResponse`](internal/api/preflight/models.go:7): metadata statistics (`metadata_attempts`, `metadata_successes`, `metadata_failures`, `metadata_fallbacks`).
   - [`PreflightResponse`](internal/api/preflight/models.go:14): contains `valid`, `error`, `unique_video_count`, `total_resource_count`, `estimated_work`, and `stats`.

3. **Preflight Service (`[`internal/youtube/preflight/service.go`](internal/youtube/preflight/service.go:1)` & `[`internal/youtube/preflight/model.go`](internal/youtube/preflight/model.go:1)`)**:
   - Interface [`preflight.Service`](internal/youtube/preflight/model.go:27) defines `Check(ctx context.Context, req Request) (Result, error)`.
   - Service implementation handles resource resolution, cache checking, estimation, and statistics collection cleanly within the domain layer.

## Architecture Guidelines

- The API layer ([[`internal/api/preflight/handler.go`](internal/api/preflight/handler.go:1)]) must only translate the preflight service result into its response format.
- It must not:
  - build execution plans;
  - inspect caches;
  - estimate metadata;
  - estimate subtitles;
  - deduplicate video IDs.
- Return the service result directly or map it with minimal translation.

## Actionable Migration Steps

1. **Verify Service Integration**: Confirm that [`internal/api/preflight/handler.go`](internal/api/preflight/handler.go:1) uses the domain [`preflight.Service`](internal/youtube/preflight/model.go:27) interface correctly without performing domain-level business logic.
2. **Review and Update Tests**: Ensure [`internal/api/preflight/handler_test.go`](internal/api/preflight/handler_test.go:1) comprehensively tests success and error conditions against the updated [`preflight.Service`](internal/youtube/preflight/model.go:27) interface.
3. **Validate End-to-End Flow**: Run existing unit and integration tests to verify the preflight API migration works correctly.
