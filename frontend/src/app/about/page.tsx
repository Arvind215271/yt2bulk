'use client';

import React from 'react';
import Header from '@/components/Header';
import { theme } from '@/theme/theme';

export default function AboutPage() {
  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <div style={{ width: '100%', maxWidth: '900px', margin: '0 auto', padding: '1.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        
        <div>
          <h1 style={{ fontSize: '1.8rem', fontWeight: 800, marginBottom: '0.75rem', letterSpacing: '-0.02em' }}>About This Tool</h1>
          <p style={{ color: theme.colors.textMuted, fontSize: '1.1rem', lineHeight: '1.6' }}>
            This application helps you pull details, descriptions, subtitles, and transcripts from YouTube videos and playlists quickly.
          </p>
        </div>

        <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700 }}>How It Works</h2>
          
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', color: theme.colors.textMuted, fontSize: '0.95rem', lineHeight: '1.6' }}>
            <p>
              <strong style={{ color: theme.colors.textMain }}>Backend Speed:</strong> Built using Go for fast processing and reliable data fetching.
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>Subtitle Selection:</strong> Pick and choose manual or automatic subtitle tracks across any video or playlist.
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>Easy Export:</strong> Save your data as JSON, CSV, Markdown, YAML, XML, or Excel spreadsheets.
            </p>
            <p>
              <strong style={{ color: theme.colors.textMain }}>Playlist Links:</strong> Quickly grab all video links from any YouTube playlist.
            </p>
          </div>
        </div>

        <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700 }}>Pages</h2>
          <ul style={{ paddingLeft: '1.25rem', color: theme.colors.textMuted, fontSize: '0.95rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', lineHeight: '1.5' }}>
            <li><strong style={{ color: theme.colors.textMain }}>Home (`/`)</strong>: Main entry point.</li>
            <li><strong style={{ color: theme.colors.textMain }}>Export (`/export`)</strong>: Download video details and transcripts.</li>
            <li><strong style={{ color: theme.colors.textMain }}>Subtitle (`/subtitle`)</strong>: Check available subtitle languages and download subtitle files.</li>
            <li><strong style={{ color: theme.colors.textMain }}>Playlist (`/playlist`)</strong>: Extract all links from a playlist.</li>
            <li><strong style={{ color: theme.colors.textMain }}>Transcript (`/transcript`)</strong>: View raw video transcripts.</li>
          </ul>
        </div>

      </div>
    </div>
  );
}
