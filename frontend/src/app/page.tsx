'use client';

import React, { useState, useMemo } from 'react';
import Header from '@/components/Header';
import { theme } from '@/theme/theme';
import Link from 'next/link';

// Data Definitions
const playlistBenchmarkData = [
  { workers: 1, time: 417.10, speedup: 1.00 },
  { workers: 4, time: 120.72, speedup: 3.45 },
  { workers: 16, time: 29.43, speedup: 14.17 },
  { workers: 32, time: 18.64, speedup: 22.38 },
  { workers: 64, time: 16.42, speedup: 25.40 },
];

const largeScaleBenchmarkData = [
  { workers: 16, time: 810.7, throughput: 6.17 },
  { workers: 32, time: 427.6, throughput: 11.69 },
  { workers: 64, time: 225.3, throughput: 22.19 },
  { workers: 128, time: 145.5, throughput: 34.36 },
  { workers: 256, time: 132.0, throughput: 37.88 },
];

export default function Home() {
  const [playlistSort, setPlaylistSort] = useState<{ key: 'workers' | 'time' | 'speedup', direction: 'asc' | 'desc' }>({ key: 'workers', direction: 'asc' });
  const [largeSort, setLargeSort] = useState<{ key: 'workers' | 'time' | 'throughput', direction: 'asc' | 'desc' }>({ key: 'workers', direction: 'asc' });

  const sortedPlaylistData = useMemo(() => {
    return [...playlistBenchmarkData].sort((a, b) => {
      const aVal = a[playlistSort.key];
      const bVal = b[playlistSort.key];
      return playlistSort.direction === 'asc' ? aVal - bVal : bVal - aVal;
    });
  }, [playlistSort]);

  const sortedLargeData = useMemo(() => {
    return [...largeScaleBenchmarkData].sort((a, b) => {
      const aVal = a[largeSort.key];
      const bVal = b[largeSort.key];
      return largeSort.direction === 'asc' ? aVal - bVal : bVal - aVal;
    });
  }, [largeSort]);

  const toggleSort = (key: any, setter: any, current: any) => {
    setter({ key, direction: current.key === key && current.direction === 'asc' ? 'desc' : 'asc' });
  };

  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <main style={{ width: '100%', maxWidth: '900px', margin: '0 auto', padding: '2.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '2rem' }}>
        
        {/* Hero Section */}
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center', gap: '1rem' }}>
          <p style={{ color: theme.colors.textMuted, fontSize: '1.15rem', maxWidth: '700px', lineHeight: '1.6' }}>
            A fast, local-first tool for extracting YouTube metadata, subtitles, and transcripts.
          </p>
        </div>

        {/* Performance Benchmark (Main Hook) */}
        <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1.5rem', boxShadow: '0 4px 20px rgba(0,0,0,0.1)' }}>
          <div>
            <h2 style={{ fontSize: '1.3rem', fontWeight: 700, color: theme.colors.primaryRed, marginBottom: '0.5rem' }}>Playlist Performance (316 Videos)</h2>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.95rem', lineHeight: '1.6', marginBottom: '1rem' }}>
              Worker scaling performance extracting video metadata and subtitle tracks concurrently (click header to sort):
            </p>
            <div style={{ overflowX: 'auto', border: `1px solid ${theme.colors.borderColor}`, borderRadius: '6px', backgroundColor: theme.colors.bg }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.95rem', color: theme.colors.textMain }}>
                <thead style={{ backgroundColor: theme.colors.primaryRed }}>
                  <tr style={{ borderBottom: `2px solid ${theme.colors.borderColor}`, borderLeft: '4px solid transparent' }}>
                    {['workers', 'time', 'speedup'].map((key) => (
                      <th key={key} style={{ padding: '1rem', textAlign: key === 'workers' ? 'left' : 'right', color: '#ffffff', fontWeight: 700, fontSize: '0.9rem', cursor: 'pointer' }} onClick={() => toggleSort(key, setPlaylistSort, playlistSort)}>
                        {key === 'workers' ? 'Workers' : key === 'time' ? 'Export Time' : 'Speedup'} {playlistSort.key === key ? (playlistSort.direction === 'asc' ? ' ▲' : ' ▼') : ' ↕'}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody style={{ fontFamily: theme.fonts.mono }}>
                  {sortedPlaylistData.map((row, idx) => {
                    const isFastest = row.workers === 64; // Based on original data
                    return (
                      <tr key={idx} style={{ borderBottom: `1px solid ${theme.colors.borderColor}`, backgroundColor: isFastest ? 'rgba(255, 30, 30, 0.15)' : 'transparent', borderLeft: `4px solid ${isFastest ? theme.colors.primaryRed : 'transparent'}` }}>
                        <td style={{ padding: '0.8rem 1rem', textAlign: 'left', fontWeight: isFastest ? 700 : 400, color: isFastest ? '#ffffff' : theme.colors.textMain }}>{row.workers}</td>
                        <td style={{ padding: '0.8rem 1rem', textAlign: 'right', fontWeight: isFastest ? 700 : 400, color: isFastest ? '#ffffff' : theme.colors.textMain }}>{row.time.toFixed(2)}s</td>
                        <td style={{ padding: '0.8rem 1rem', textAlign: 'right', fontWeight: isFastest ? 700 : 400, color: isFastest ? '#ffffff' : theme.colors.textMain }}>{row.speedup.toFixed(2)}x</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>

          <div>
            <h2 style={{ fontSize: '1.3rem', fontWeight: 700, color: theme.colors.primaryRed, marginBottom: '0.5rem' }}>Large-Scale Processing (5,000 Videos)</h2>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.95rem', lineHeight: '1.6', marginBottom: '1rem' }}>
              High-volume throughput scaling (click header to sort):
            </p>
            <div style={{ overflowX: 'auto', border: `1px solid ${theme.colors.borderColor}`, borderRadius: '6px', backgroundColor: theme.colors.bg }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.95rem', color: theme.colors.textMain }}>
                <thead style={{ backgroundColor: theme.colors.primaryRed }}>
                  <tr style={{ borderBottom: `2px solid ${theme.colors.borderColor}`, borderLeft: '4px solid transparent' }}>
                    {['workers', 'time', 'throughput'].map((key) => (
                      <th key={key} style={{ padding: '1rem', textAlign: key === 'workers' ? 'left' : 'right', color: '#ffffff', fontWeight: 700, fontSize: '0.9rem', cursor: 'pointer' }} onClick={() => toggleSort(key, setLargeSort, largeSort)}>
                        {key === 'workers' ? 'Workers' : key === 'time' ? 'Total Time' : 'Throughput'} {largeSort.key === key ? (largeSort.direction === 'asc' ? ' ▲' : ' ▼') : ' ↕'}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody style={{ fontFamily: theme.fonts.mono }}>
                  {sortedLargeData.map((row, idx) => {
                    const isFastest = row.workers === 256; // Based on original data
                    return (
                      <tr key={idx} style={{ borderBottom: `1px solid ${theme.colors.borderColor}`, backgroundColor: isFastest ? 'rgba(255, 30, 30, 0.15)' : 'transparent', borderLeft: `4px solid ${isFastest ? theme.colors.primaryRed : 'transparent'}` }}>
                        <td style={{ padding: '0.8rem 1rem', textAlign: 'left', fontWeight: isFastest ? 700 : 400, color: isFastest ? '#ffffff' : theme.colors.textMain }}>{row.workers}</td>
                        <td style={{ padding: '0.8rem 1rem', textAlign: 'right', fontWeight: isFastest ? 700 : 400, color: isFastest ? '#ffffff' : theme.colors.textMain }}>{row.time.toFixed(1)}s</td>
                        <td style={{ padding: '0.8rem 1rem', textAlign: 'right', fontWeight: isFastest ? 700 : 400, color: isFastest ? '#ffffff' : theme.colors.textMain }}>{row.throughput.toFixed(2)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.85rem', fontStyle: 'italic', marginTop: '0.75rem', lineHeight: '1.5' }}>
              Note: At higher worker counts, YouTube may rate-limit subtitle downloads, resulting in empty or missing subtitles. Using 4–64 workers is recommended; for larger playlists, lower worker counts are preferred to maintain stability.
            </p>
          </div>
        </div>

        {/* How It Works */}
        <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          <h2 style={{ fontSize: '1.2rem', fontWeight: 700 }}>How It Works</h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', color: theme.colors.textMuted, fontSize: '0.95rem', lineHeight: '1.6' }}>
            <p>
              <strong style={{ color: theme.colors.textMain }}>Backend Speed:</strong> Built using Go for ultra-fast concurrent processing and reliable yt-dlp wrapper execution.
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>Subtitle Selection:</strong> Inspect and download manual or automatic subtitle tracks across any video or playlist.
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>Easy Export:</strong> Save your data as JSON, CSV, Markdown, YAML, XML, or Excel spreadsheets with customized fields.
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>Playlist Links & Preflight:</strong> Quickly grab all video links or run detailed accounting and estimation checks before heavy processing.
            </p>
          </div>
        </div>

        {/* Why Run Locally */}
        <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          <h2 style={{ fontSize: '1.2rem', fontWeight: 700 }}>Why Run Locally?</h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem', color: theme.colors.textMuted, fontSize: '0.95rem', lineHeight: '1.6' }}>
            <p>
              Running <strong style={{ color: theme.colors.textMain }}>YT2BULK</strong> locally gives you full, unrestricted access to YouTube data pipelines without the constraints of hosted cloud versions:
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>1. Zero Export & Request Limits:</strong> Export massive playlists, fetch extensive subtitles, and process bulk data without artificial rate caps or concurrency limits.
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>2. Full Privacy & Control:</strong> Your queries, exports, and data stay entirely on your machine with zero external tracking or logging.
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>3. Persistent Workers & Native Concurrency:</strong> Leverage local Python workers and Go routines for maximum speed and zero cold starts.
            </p>
          </div>
        </div>

        {/* Available Pages */}
        <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          <h2 style={{ fontSize: '1.2rem', fontWeight: 700 }}>Available Tools & Pages</h2>
          <ul style={{ paddingLeft: '1.25rem', color: theme.colors.textMuted, fontSize: '0.95rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', lineHeight: '1.5' }}>
            <li><strong style={{ color: theme.colors.textMain }}><Link href="/export">Export (`/export`)</Link></strong>: Comprehensive metadata and transcript extraction in multiple formats.</li>
            <li><strong style={{ color: theme.colors.textMain }}><Link href="/playlist">Playlist (`/playlist`)</Link></strong>: Extract all individual video links and URLs from any playlist.</li>
            <li><strong style={{ color: theme.colors.textMain }}><Link href="/subtitle">Subtitle (`/subtitle`)</Link></strong>: Browse and download manual or auto-generated subtitle tracks.</li>
            <li><strong style={{ color: theme.colors.textMain }}><Link href="/preflight">Preflight (`/preflight`)</Link></strong>: Pre-flight analysis, accounting, and resource estimation for playlist pipelines.</li>
          </ul>
        </div>

      </main>
    </div>
  );
}
