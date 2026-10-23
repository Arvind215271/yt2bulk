'use client';

import React from 'react';
import Header from '@/components/Header';
import { theme } from '@/theme/theme';
import Link from 'next/link';

export default function Home() {
  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <main style={{ width: '100%', maxWidth: '900px', margin: '0 auto', padding: '2.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '2rem' }}>
        
        {/* Hero Section */}
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center', gap: '1rem' }}>
          <h1 style={{ fontSize: '2.5rem', fontWeight: 800, letterSpacing: '-0.03em', color: theme.colors.primaryRed }}>
            YouTube Data Tool
          </h1>
          <p style={{ color: theme.colors.textMuted, fontSize: '1.15rem', maxWidth: '700px', lineHeight: '1.6' }}>
            A powerful, high-performance local tool to pull video details, descriptions, subtitles, and metadata from YouTube links or playlists with zero limits.
          </p>
          <div style={{ display: 'flex', flexWrap: 'wrap', justifyContent: 'center', gap: '1rem', marginTop: '1rem' }}>
            <Link
              href="/export"
              style={{ backgroundColor: theme.colors.primaryRed, color: '#ffffff', fontWeight: 600, padding: '0.75rem 1.5rem', borderRadius: theme.spacing.borderRadius, textDecoration: 'none', fontSize: '0.95rem' }}
            >
              Export Tool
            </Link>
            <Link
              href="/playlist"
              style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, color: theme.colors.textMain, fontWeight: 600, padding: '0.75rem 1.5rem', borderRadius: theme.spacing.borderRadius, textDecoration: 'none', fontSize: '0.95rem' }}
            >
              Playlist Links
            </Link>
            <Link
              href="/subtitle"
              style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, color: theme.colors.textMain, fontWeight: 600, padding: '0.75rem 1.5rem', borderRadius: theme.spacing.borderRadius, textDecoration: 'none', fontSize: '0.95rem' }}
            >
              Subtitles
            </Link>
            <Link
              href="/preflight"
              style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, color: theme.colors.textMain, fontWeight: 600, padding: '0.75rem 1.5rem', borderRadius: theme.spacing.borderRadius, textDecoration: 'none', fontSize: '0.95rem' }}
            >
              Preflight Check
            </Link>
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
