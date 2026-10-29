'use client';

import React, { useState, useMemo, useRef } from 'react';
import Header from '@/components/Header';
import { theme } from '@/theme/theme';
import { fetchChannelVideos, ChannelVideosResponse, VideoEntry } from '@/services/channel';

function formatDuration(seconds: number): string {
  if (!seconds || isNaN(seconds)) return '0:00';
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  const pad = (n: number) => n.toString().padStart(2, '0');
  if (hrs > 0) {
    return `${hrs}:${pad(mins)}:${pad(secs)}`;
  }
  return `${mins}:${pad(secs)}`;
}

export default function ChannelPage() {
  const [url, setUrl] = useState('');
  const [provider, setProvider] = useState('ytdlp');
  const [limitMode, setLimitMode] = useState<'all' | 'first' | 'range'>('all');
  const [limitCount, setLimitCount] = useState<number>(50);
  const [rangeStart, setRangeStart] = useState<number>(1);
  const [rangeEnd, setRangeEnd] = useState<number>(50);

  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ChannelVideosResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Search & Selection state
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedIds, setSelectedIds] = useState<Record<string, boolean>>({});

  const abortControllerRef = useRef<AbortController | null>(null);

  const handleCancel = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setLoading(false);
    setError('Channel video extraction cancelled by user.');
  };

  const handleFetch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim()) {
      alert('Please enter a YouTube channel URL or ID.');
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);
    setSelectedIds({});

    const controller = new AbortController();
    abortControllerRef.current = controller;

    let effectiveLimit = 0;
    if (limitMode === 'first') {
      effectiveLimit = limitCount;
    } else if (limitMode === 'range') {
      effectiveLimit = rangeEnd;
    }

    try {
      const resp = await fetchChannelVideos({
        url: url.trim(),
        provider,
        output: 'both',
        limit_mode: limitMode,
        limit: effectiveLimit,
        range_start: rangeStart,
        range_end: rangeEnd
      }, controller.signal);

      setResult({
        ...resp,
        videos: resp.videos,
        total: resp.total
      });
    } catch (err: any) {
      if (err.name === 'AbortError' || err.message?.includes('aborted')) {
        setError('Channel extraction cancelled.');
      } else {
        setError(err.message || 'Failed to fetch channel videos');
      }
    } finally {
      setLoading(false);
      abortControllerRef.current = null;
    }
  };

  const filteredVideos = useMemo(() => {
    if (!result || !result.videos) return [];
    if (!searchQuery.trim()) return result.videos;
    const q = searchQuery.toLowerCase();
    return result.videos.filter(v => 
      (v.title && v.title.toLowerCase().includes(q)) ||
      (v.id && v.id.toLowerCase().includes(q)) ||
      (v.url && v.url.toLowerCase().includes(q))
    );
  }, [result, searchQuery]);

  const handleSelectAll = (checked: boolean) => {
    const newSelected: Record<string, boolean> = {};
    if (checked) {
      filteredVideos.forEach(v => {
        newSelected[v.id || v.url] = true;
      });
    }
    setSelectedIds(newSelected);
  };

  const handleToggleSelect = (id: string) => {
    setSelectedIds(prev => ({
      ...prev,
      [id]: !prev[id]
    }));
  };

  const selectedCount = Object.values(selectedIds).filter(Boolean).length;

  const handleCopySelected = () => {
    if (!result) return;
    const items = result.videos.filter(v => selectedIds[v.id || v.url]);
    if (items.length === 0) {
      alert('No videos selected!');
      return;
    }
    const text = items.map(v => v.url || `https://www.youtube.com/watch?v=${v.id}`).join('\n');
    navigator.clipboard.writeText(text);
    alert(`Copied ${items.length} selected video links to clipboard!`);
  };

  const handleCopyAll = () => {
    if (!result || !result.videos) return;
    const text = result.videos.map(v => v.url || `https://www.youtube.com/watch?v=${v.id}`).join('\n');
    navigator.clipboard.writeText(text);
    alert(`Copied all ${result.videos.length} video links to clipboard!`);
  };

  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <div style={{ width: '100%', maxWidth: '1200px', margin: '0 auto', padding: '1.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        
        <div>
          <h1 style={{ fontSize: '1.8rem', fontWeight: 800, marginBottom: '0.5rem', letterSpacing: '-0.02em' }}>YouTube Channel Export & Video Browser</h1>
          <p style={{ color: theme.colors.textMuted, fontSize: '1.05rem' }}>
            Discover, search, filter, and export all videos from any YouTube channel instantly with metadata and custom range limits.
          </p>
        </div>

        <form onSubmit={handleFetch} style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', width: '100%' }}>
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem' }}>
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem' }}>Channel Input & Extraction Controls</h2>
            
            <div style={{ marginBottom: '1.25rem' }}>
              <label style={{ display: 'block', fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.5rem', color: theme.colors.textMuted }}>
                Channel URL or ID (@handle, /channel/, /c/, /user/)
              </label>
              <input
                type="text"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://www.youtube.com/@ChannelName or @handle"
                style={{
                  width: '100%',
                  backgroundColor: theme.colors.bg,
                  border: `1px solid ${theme.colors.borderColor}`,
                  borderRadius: '6px',
                  padding: '0.9rem 1rem',
                  fontSize: '0.95rem',
                  color: theme.colors.textMain,
                  fontFamily: theme.fonts.mono
                }}
              />
            </div>

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
                  Extraction Limit Mode
                </label>
                <select
                  value={limitMode}
                  onChange={(e) => setLimitMode(e.target.value as any)}
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
                  <option value="all">All Videos</option>
                  <option value="first">First X Videos</option>
                  <option value="range">Position Range (e.g. 1-100)</option>
                </select>
              </div>

              {limitMode === 'first' && (
                <div>
                  <label style={{ display: 'block', fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.5rem', color: theme.colors.textMuted }}>
                    Limit Count (X)
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={10000}
                    value={limitCount}
                    onChange={(e) => setLimitCount(parseInt(e.target.value) || 50)}
                    style={{
                      width: '100%',
                      backgroundColor: theme.colors.bg,
                      border: `1px solid ${theme.colors.borderColor}`,
                      borderRadius: '6px',
                      padding: '0.6rem 1rem',
                      fontSize: '0.95rem',
                      color: theme.colors.textMain
                    }}
                  />
                </div>
              )}

              {limitMode === 'range' && (
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <div style={{ flex: 1 }}>
                    <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, marginBottom: '0.5rem', color: theme.colors.textMuted }}>Start #</label>
                    <input
                      type="number"
                      min={1}
                      value={rangeStart}
                      onChange={(e) => setRangeStart(parseInt(e.target.value) || 1)}
                      style={{
                        width: '100%',
                        backgroundColor: theme.colors.bg,
                        border: `1px solid ${theme.colors.borderColor}`,
                        borderRadius: '6px',
                        padding: '0.6rem 0.8rem',
                        fontSize: '0.9rem',
                        color: theme.colors.textMain
                      }}
                    />
                  </div>
                  <div style={{ flex: 1 }}>
                    <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, marginBottom: '0.5rem', color: theme.colors.textMuted }}>End #</label>
                    <input
                      type="number"
                      min={1}
                      value={rangeEnd}
                      onChange={(e) => setRangeEnd(parseInt(e.target.value) || 50)}
                      style={{
                        width: '100%',
                        backgroundColor: theme.colors.bg,
                        border: `1px solid ${theme.colors.borderColor}`,
                        borderRadius: '6px',
                        padding: '0.6rem 0.8rem',
                        fontSize: '0.9rem',
                        color: theme.colors.textMain
                      }}
                    />
                  </div>
                </div>
              )}
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
                  padding: '0.75rem 1.25rem',
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
                <span>Extract Channel Videos</span>
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
                    padding: '0.75rem 1.25rem',
                    fontWeight: 700,
                    fontSize: '0.9rem',
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
                <h2 style={{ fontSize: '1.2rem', fontWeight: 700, marginBottom: '0.25rem' }}>Channel Videos & Listing</h2>
                <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem' }}>
                  Channel ID: <strong style={{ color: theme.colors.textMain }}>{result.channel_id}</strong> | Total Found: <strong style={{ color: theme.colors.primaryRed }}>{result.total}</strong> | Selected: <strong style={{ color: theme.colors.primaryRed }}>{selectedCount}</strong>
                </p>
              </div>

              <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
                <button
                  onClick={handleCopySelected}
                  style={{
                    backgroundColor: theme.colors.primaryRed,
                    color: '#ffffff',
                    border: 'none',
                    borderRadius: '6px',
                    padding: '0.5rem 1rem',
                    fontSize: '0.85rem',
                    fontWeight: 600,
                    cursor: 'pointer'
                  }}
                >
                  Copy Selected Links ({selectedCount})
                </button>
                <button
                  onClick={() => {
                    const selectedVideos = result.videos.filter(v => selectedIds[v.id || v.url]);
                    if (selectedVideos.length === 0) {
                      alert('No videos selected!');
                      return;
                    }
                    const urls = selectedVideos.map(v => v.url || `https://www.youtube.com/watch?v=${v.id}`);
                    localStorage.setItem('export_inputs', urls.join('\n'));
                    window.location.href = '/export';
                  }}
                  style={{
                    backgroundColor: theme.colors.bg,
                    color: theme.colors.textMain,
                    border: `1px solid ${theme.colors.borderColor}`,
                    borderRadius: '6px',
                    padding: '0.5rem 1rem',
                    fontSize: '0.85rem',
                    fontWeight: 600,
                    cursor: 'pointer'
                  }}
                >
                  Export Selected ({selectedCount})
                </button>
                <button
                  onClick={handleCopyAll}
                  style={{
                    backgroundColor: theme.colors.bg,
                    color: theme.colors.textMain,
                    border: `1px solid ${theme.colors.borderColor}`,
                    borderRadius: '6px',
                    padding: '0.5rem 1rem',
                    fontSize: '0.85rem',
                    fontWeight: 600,
                    cursor: 'pointer'
                  }}
                >
                  Copy All Links
                </button>
              </div>
            </div>

            {/* Search Filter Bar */}
            <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
              <input
                type="text"
                placeholder="Search videos by title, ID, or URL..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                style={{
                  flex: 1,
                  backgroundColor: theme.colors.bg,
                  border: `1px solid ${theme.colors.borderColor}`,
                  borderRadius: '6px',
                  padding: '0.6rem 1rem',
                  fontSize: '0.9rem',
                  color: theme.colors.textMain
                }}
              />
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery('')}
                  style={{
                    backgroundColor: 'transparent',
                    border: `1px solid ${theme.colors.borderColor}`,
                    color: theme.colors.textMuted,
                    borderRadius: '6px',
                    padding: '0.6rem 1rem',
                    fontSize: '0.85rem',
                    cursor: 'pointer'
                  }}
                >
                  Clear Search
                </button>
              )}
            </div>

            {/* Videos Table */}
            <div style={{
              maxHeight: '550px',
              overflowY: 'auto',
              border: `1px solid ${theme.colors.borderColor}`,
              borderRadius: '6px',
              backgroundColor: theme.colors.bg
            }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.9rem' }}>
                <thead>
                  <tr style={{ borderBottom: `1px solid ${theme.colors.borderColor}`, backgroundColor: 'rgba(255,255,255,0.02)', position: 'sticky', top: 0, zIndex: 10 }}>
                    <th style={{ padding: '0.8rem 1rem', width: '40px' }}>
                      <input
                        type="checkbox"
                        onChange={(e) => handleSelectAll(e.target.checked)}
                        checked={filteredVideos.length > 0 && filteredVideos.every(v => selectedIds[v.id || v.url])}
                        style={{ cursor: 'pointer' }}
                      />
                    </th>
                    <th style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontWeight: 600, width: '60px' }}>#</th>
                    <th style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontWeight: 600, width: '100px' }}>Thumbnail</th>
                    <th style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontWeight: 600 }}>Title</th>
                    <th style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontWeight: 600, width: '90px' }}>Duration</th>
                    <th style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontWeight: 600, width: '120px' }}>Video ID</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredVideos.length === 0 ? (
                    <tr>
                      <td colSpan={6} style={{ padding: '2rem', textAlign: 'center', color: theme.colors.textMuted }}>
                        No matching videos found.
                      </td>
                    </tr>
                  ) : (
                    filteredVideos.map((vid, idx) => {
                      const identifier = vid.id || vid.url;
                      const isSelected = !!selectedIds[identifier];
                      const thumbUrl = vid.id ? `https://i.ytimg.com/vi/${vid.id}/hqdefault.jpg` : '';

                      return (
                        <tr key={idx} style={{ borderBottom: `1px solid ${theme.colors.borderColor}`, backgroundColor: isSelected ? 'rgba(255, 30, 30, 0.08)' : 'transparent' }}>
                          <td style={{ padding: '0.8rem 1rem' }}>
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={() => handleToggleSelect(identifier)}
                              style={{ cursor: 'pointer' }}
                            />
                          </td>
                          <td style={{ padding: '0.8rem 1rem', color: theme.colors.textMuted, fontFamily: theme.fonts.mono }}>{vid.position}</td>
                          <td style={{ padding: '0.5rem 1rem' }}>
                            {thumbUrl ? (
                              <img src={thumbUrl} alt="" style={{ width: '80px', height: '45px', objectFit: 'cover', borderRadius: '4px', border: `1px solid ${theme.colors.borderColor}` }} />
                            ) : (
                              <div style={{ width: '80px', height: '45px', backgroundColor: theme.colors.panelBg, borderRadius: '4px', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.75rem', color: theme.colors.textMuted }}>N/A</div>
                            )}
                          </td>
                          <td style={{ padding: '0.8rem 1rem', color: theme.colors.textMain }}>
                            {vid.url ? (
                              <a href={vid.url} target="_blank" rel="noopener noreferrer" style={{ color: theme.colors.textMain, textDecoration: 'none', fontWeight: 500 }} title={vid.title}>
                                {vid.title || vid.url}
                              </a>
                            ) : (
                              <span>{vid.title || 'Untitled'}</span>
                            )}
                          </td>
                          <td style={{ padding: '0.8rem 1rem', fontFamily: theme.fonts.mono, color: theme.colors.textMuted, fontSize: '0.85rem' }}>
                            {formatDuration(vid.duration)}
                          </td>
                          <td style={{ padding: '0.8rem 1rem', fontFamily: theme.fonts.mono, color: theme.colors.textMuted, fontSize: '0.85rem' }}>
                            {vid.id || '-'}
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>

          </div>
        )}

      </div>
    </div>
  );
}
