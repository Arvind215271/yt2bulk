ossible# `yt-dlp` Player Client Compatibility

This document outlines the results of testing various `yt-dlp` `--extractor-args` "youtube:player-client=..." options regarding subtitle availability and successful video format extraction.

## Compatibility Table

| Player Client | Subtitle Availability | Successful Extraction (Exit 0) | Subtitle Export Supported |
| :--- | :--- | :--- | :--- |
| `android` | Yes | **Yes** | **Yes** |
| `android_vr` | Yes | **Yes** | **Yes** |
| `visionos` | Yes | **Yes** | **Yes** |
| `tv_samsung` | Yes | **Yes** | **Yes** |
| `default` | Yes | **Yes** | **Yes** |
| `all` | Yes | **Yes** | **Yes** |
| `web` | Yes | No | No |
| `web_safari` | Yes | No | No |
| `mweb` | Yes | No | No |
| `ios` | Yes | No | No |
| `web_embedded` | No | No | No |
| `web_music` | No | No | No |
| `web_creator` | No | No | No |
| `tv` | No | No | No |
| `tv_simply` | No | No | No |
| `tv_downgraded` | No | No | No |

## Key Findings

*   **Successful Clients:** The clients `android`, `android_vr`, `visionos`, `tv_samsung`, `default`, and `all` consistently provided both subtitles and video format information, enabling successful extraction and subtitle export.
*   **Format Extraction Failures:** While `web`, `web_safari`, `mweb`, and `ios` could detect available subtitles, they failed the overall extraction process because they could not retrieve valid video formats (likely due to the missing PO Token/challenge requirements).
*   **Recommendations:** For reliable subtitle extraction and export, utilize `android`, `visionos`, or `tv_samsung` as the `player-client`. These consistently allow for valid subtitle downloads without requiring complex manual cookie or PO token configuration.
