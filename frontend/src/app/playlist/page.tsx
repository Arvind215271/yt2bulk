'use client';

import React, { useState } from 'react';
import Header from '@/components/Header';
import { theme } from '@/theme/theme';

interface VideoEntry {
  id?: string;
  url?: string;
  position: number;
  added_at?: string;
}

interface PlaylistResponse {
  playlist_id: string;
  total: number;
  videos: VideoEntry[];
}

export default function PlaylistPage() {
  const [url, setUrl] = useState('');
  const [provider, setProvider] = useState('ytdlp');
  const [playerClient, setPlayerClient] = useState('default');
  const [output, setOutput] = useState('both');
  
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<PlaylistResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const abortControllerRef = React.useRef<AbortController | null>(null);

  const handleCancel = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setLoading(false);
    setError('Request cancelled by user.');
  };

  const handleExpand = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim()) {
      alert('Please enter a YouTube playlist URL or ID.');
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const res = await fetch(`${apiUrl}/playlist/videos`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: url.trim(), provider, player_client: playerClient, output }),
        signal: controller.signal
      });

      const text = await res.text();
      if (!res.ok) {
        throw new Error(text || 'Failed to expand playlist');
      }

      const parsed = JSON.parse(text);
      setResult(parsed);
    } catch (err: any) {
      if (err.name === 'AbortError' || err.message?.includes('aborted')) {
        setError('Playlist expansion cancelled by user.');
      } else {
        setError(err.message);
      }
    } finally {
      setLoading(false);
      abortControllerRef.current = null;
    }
  };

  const handleCopyAll = () => {
    if (!result || !result.videos) return;
    const lines = result.videos.map(v => v.url || `https://www.youtube.com/watch?v=${v.id}`).join('\n');
    navigator.clipboard.writeText(lines);
    alert('Copied all video links to clipboard!');
  };

  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <div style={{ width: '100%', maxWidth: '1100px', margin: '0 auto', padding: '1.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        
        <div>
          <h1 style={{ fontSize: '1.8rem', fontWeight: 800, marginBottom: '0.5rem', letterSpacing: '-0.02em' }}>Playlist Video Expander</h1>
          <p style={{ color: theme.colors.textMuted, fontSize: '1.05rem' }}>
            Extract and list all video IDs, URLs, and positions from any YouTube playlist instantly using high-performance backend pipelines.
          </p>
        </div>

        <form onSubmit={handleExpand} style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', width: '100%' }}>
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem' }}>
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem' }}>Playlist Input & Configuration</h2>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem', marginBottom: '1.25rem' }}>
              Enter the YouTube playlist URL or ID:
            </p>
            <input
              type="text"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://www.youtube.com/playlist?list=..."
              style={{
                width: '100%',
                backgroundColor: theme.colors.bg,
                border: `1px solid ${theme.colors.borderColor}`,
                borderRadius: '6px',
                padding: '0.9rem 1rem',
                fontSize: '0.95rem',
                color: theme.colors.textMain,
                fontFamily: theme.fonts.mono,
                marginBottom: '1.5rem'
              }}
            />

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1.5rem', marginBottom: '1.5rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.5rem', color: theme.colors.textMuted }}>
                  Provider
                </label>
                <select
                  value={provider}
                  onChange={(e) => setProvider(e.target.value)}
                  style={{
                    width: '100%',
                    backgroundColor: theme.colors.bg,
                    border: `1px solid ${theme.colors.borderColor}`,
                    borderRadius: '6px',
                    padding: '0.6rem 1rem',
                    fontSize: '0.95rem',
                    color: theme.colors.textMain
                  }}
                >
                  <option value="ytdlp">yt-dlp (Recommended)</option>
                  <option value="api">YouTube API v3</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.5rem', color: theme.colors.textMuted }}>
                  Player Client
                </label>
                <select
                  value={playerClient}
                  onChange={(e) => setPlayerClient(e.target.value)}
                  style={{
                    width: '100%',
                    backgroundColor: theme.colors.bg,
                    border: `1px solid ${theme.colors.borderColor}`,
                    borderRadius: '6px',
                    padding: '0.6rem 1rem',
                    fontSize: '0.95rem',
                    color: theme.colors.textMain
                  }}
                >
                  <option value="default">Default</option>
                  <option value="android">Android</option>
                  <option value="android_vr">Android VR</option>
                  <option value="visionos">VisionOS</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.5rem', color: theme.colors.textMuted }}>
                  Output Mode
                </label>
                <select
                  value={output}
                  onChange={(e) => setOutput(e.target.value)}
                  style={{
                    width: '100%',
                    backgroundColor: theme.colors.bg,
                    border: `1px solid ${theme.colors.borderColor}`,
                    borderRadius: '6px',
                    padding: '0.6rem 1rem',
                    fontSize: '0.95rem',
                    color: theme.colors.textMain
                  }}
                >
                  <option value="both">Both ID & URL</option>
                  <option value="url">URL Only</option>
                  <option value="id">ID Only</option>
                </select>
              </div>
            </div>

            <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
              <button
                type="submit"
                disabled={loading}
                style={{
                  backgroundColor: theme.colors.primaryRed,
                  color: '#ffffff',
                  border: 'none',
                  borderRadius: '6px',
                  padding: '0.75rem 1rem',
                  fontWeight: 700,
                  fontSize: '0.9rem',
                  cursor: 'pointer',
                  opacity: loading ? 0.7 : 1,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.75rem',
                  boxShadow: `0 4px 14px ${theme.colors.accentGlow}`
                }}
              >
                {loading && <div style={{ width: '1rem', height: '1rem', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', borderRadius: '50%', animation: 'spin 0.8s infinite linear' }}></div>}
                <span>Expand Playlist Videos</span>
              </button>

              {loading && (
                <button
                  type="button"
                  onClick={handleCancel}
                  style={{
                    backgroundColor: 'transparent',
                    color: theme.colors.primaryRed,
                    border: `1px solid ${theme.colors.primaryRed}`,
                    borderRadius: '6px',
                    padding: '0.9rem 1.5rem',
                    fontWeight: 700,
                    fontSize: '1rem',
                    cursor: 'pointer'
                  }}
                >
                  Cancel
                </button>
              )}
            </div>
          </div>
        </form>

        {error && (
          <div style={{ backgroundColor: 'rgba(255, 30, 30, 0.1)', border: `1px solid ${theme.colors.primaryRed}`, borderRadius: '6px', padding: '1rem', color: theme.colors.primaryRed }}>
            <strong>Error:</strong> {error}
          </div>
        )}

        {result && (
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
              <div>
                <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.25rem' }}>Playlist Results</h2>
                <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem' }}>
                  Playlist ID: <strong style={{ color: theme.colors.textMain }}>{result.playlist_id}</strong> | Total Videos: <strong style={{ color: theme.colors.primaryRed }}>{result.total}</strong>
                </p>
              </div>
              <button
                onClick={handleCopyAll}
                style={{
                  backgroundColor: theme.colors.bg,
                  color: theme.colors.textMain,
                  border: `1px solid ${theme.colors.borderColor}`,
                  borderRadius: '6px',
                  padding: '0.6rem 1.2rem',
                  fontSize: '0.85rem',
                  fontWeight: 600,
                  cursor: 'pointer'
                }}
              >
                Copy All Links
              </button>
            </div>

            <div style={{
              maxHeight: '450px',
              overflowY: 'auto',
              border: `1px solid ${theme.colors.borderColor}`,
              borderRadius: '6px',
              backgroundColor: theme.colors.bg
            }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.9rem' }}>
                <thead>
                  <tr style={{ borderBottom: `1px solid ${theme.colors.borderColor}`, backgroundColor: 'rgba(255,255,255,0.02)' }}>
                    <th style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontWeight: 600, width: '80px' }}>#</th>
                    <th style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontWeight: 600 }}>Video ID</th>
                    <th style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontWeight: 600 }}>URL</th>
                  </tr>
                </thead>
                <tbody>
                  {result.videos.map((vid, idx) => (
                    <tr key={idx} style={{ borderBottom: `1px solid ${theme.colors.borderColor}` }}>
                      <td style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontFamily: theme.fonts.mono }}>{vid.position}</td>
                      <td style={{ padding: '0.8rem 1rem', fontFamily: theme.fonts.mono, color: theme.colors.textMain }}>{vid.id || '-'}</td>
                      <td style={{ padding: '0.8rem 1rem' }}>
                        {vid.url ? (
                          <a href={vid.url} target="_blank" rel="noopener noreferrer" style={{ color: theme.colors.primaryRed, textDecoration: 'none' }}>
                            {vid.url}
                          </a>
                        ) : (
                          '-'
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}
