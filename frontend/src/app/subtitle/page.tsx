'use client';

import React, { useState, useMemo } from 'react';
import Header from '@/components/Header';
import { theme } from '@/theme/theme';
import { useError } from '@/components/ErrorContext';

interface SubtitleTrack {
  languageCode: string;
  languageName: string;
  formats: string[];
}

interface VideoInfo {
  id: string;
  title?: string;
  channel_title?: string;
  subtitle_metadata?: {
    manual?: SubtitleTrack[];
    automatic?: SubtitleTrack[];
  };
}

interface SubtitlePreference {
  language: string;
  type: 'manual' | 'automatic';
}

export default function SubtitlePage() {
  const { addError } = useError();
  const [inputs, setInputs] = useState('');
  const [loading, setLoading] = useState(false);
  const [rawResponseData, setRawResponseData] = useState<any>(null);

  const [searchQuery, setSearchQuery] = useState('');
  const [selectedPrefs, setSelectedPrefs] = useState<SubtitlePreference[]>([]);
  const [format, setFormat] = useState('vtt');
  const [downloading, setDownloading] = useState(false);

  // Drag and drop state
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);

  const abortControllerRef = React.useRef<AbortController | null>(null);

  const handleCancel = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setLoading(false);
    setDownloading(false);
  };

  const handleFetchOptions = async (e: React.FormEvent) => {
    e.preventDefault();
    const list = inputs.split('\n').map(s => s.trim()).filter(Boolean);
    if (list.length === 0) {
      addError('Please enter at least one YouTube video or playlist link/ID.');
      return;
    }

    setLoading(true);
    setRawResponseData(null);
    setSelectedPrefs([]);

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const res = await fetch(`${apiUrl}/subtitle/options`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ inputs: list }),
        signal: controller.signal
      });

      const text = await res.text();
      if (!res.ok) {
        throw new Error(text || 'Failed to fetch subtitle options');
      }

      const parsed = JSON.parse(text);
      setRawResponseData(parsed);
    } catch (err: any) {
      if (err.name !== 'AbortError' && !err.message?.includes('aborted')) {
        addError(err.message, { context: 'subtitle_options' });
      }
    } finally {
      setLoading(false);
      abortControllerRef.current = null;
    }
  };

  const allVideos = useMemo(() => {
    if (!rawResponseData) return [];
    const vids: VideoInfo[] = [];
    const items = Array.isArray(rawResponseData) ? rawResponseData : (rawResponseData.resources || [rawResponseData]);

    for (const item of items) {
      if (!item) continue;
      if (item.type === 'video' && item.video) {
        vids.push(item.video);
      } else if (item.subtitle_metadata) {
        vids.push(item as VideoInfo);
      } else if (item.type === 'playlist' && item.playlist && item.playlist.videos) {
        vids.push(...item.playlist.videos);
      } else if (item.videos) {
        vids.push(...item.videos);
      }
    }
    return vids;
  }, [rawResponseData]);

  const coveredVideoIds = useMemo(() => {
    const covered = new Set<string>();
    if (selectedPrefs.length === 0 || allVideos.length === 0) return covered;

    for (const vid of allVideos) {
      const manualCodes = new Set((vid.subtitle_metadata?.manual || []).map(t => t.languageCode));
      const autoCodes = new Set((vid.subtitle_metadata?.automatic || []).map(t => t.languageCode));

      for (const pref of selectedPrefs) {
        if (pref.type === 'manual' && manualCodes.has(pref.language)) {
          covered.add(vid.id);
          break;
        }
        if (pref.type === 'automatic' && autoCodes.has(pref.language)) {
          covered.add(vid.id);
          break;
        }
      }
    }
    return covered;
  }, [allVideos, selectedPrefs]);

  const subtitleStats = useMemo(() => {
    const totalVideos = allVideos.length;
    if (totalVideos === 0) return [];

    const statsMap = new Map<string, {
      languageCode: string;
      languageName: string;
      type: 'manual' | 'automatic';
      count: number;
      remainingCount: number;
      totalVideos: number;
      videoIds: string[];
    }>();

    for (const vid of allVideos) {
      const manualTracks = vid.subtitle_metadata?.manual || [];
      const autoTracks = vid.subtitle_metadata?.automatic || [];
      const isVidCovered = coveredVideoIds.has(vid.id);

      for (const track of manualTracks) {
        if (!track.languageCode) continue;
        const key = `manual_${track.languageCode}`;
        if (!statsMap.has(key)) {
          statsMap.set(key, {
            languageCode: track.languageCode,
            languageName: track.languageName || track.languageCode,
            type: 'manual',
            count: 0,
            remainingCount: 0,
            totalVideos,
            videoIds: []
          });
        }
        const stat = statsMap.get(key)!;
        stat.count++;
        if (!isVidCovered) {
          stat.remainingCount++;
        }
        stat.videoIds.push(vid.id);
      }

      for (const track of autoTracks) {
        if (!track.languageCode) continue;
        const key = `automatic_${track.languageCode}`;
        if (!statsMap.has(key)) {
          statsMap.set(key, {
            languageCode: track.languageCode,
            languageName: track.languageName || track.languageCode,
            type: 'automatic',
            count: 0,
            remainingCount: 0,
            totalVideos,
            videoIds: []
          });
        }
        const stat = statsMap.get(key)!;
        stat.count++;
        if (!isVidCovered) {
          stat.remainingCount++;
        }
        stat.videoIds.push(vid.id);
      }
    }

    const statsList = Array.from(statsMap.values());
    statsList.sort((a, b) => b.remainingCount - a.remainingCount || b.count - a.count);
    return statsList;
  }, [allVideos, coveredVideoIds]);

  const filteredStats = useMemo(() => {
    if (!searchQuery.trim()) return subtitleStats;
    const q = searchQuery.toLowerCase();
    return subtitleStats.filter(s => 
      s.languageCode.toLowerCase().includes(q) || 
      s.languageName.toLowerCase().includes(q) ||
      s.type.toLowerCase().includes(q)
    );
  }, [subtitleStats, searchQuery]);

  const coverageMetrics = useMemo(() => {
    const total = allVideos.length;
    const coveredCount = coveredVideoIds.size;
    return {
      coveredVideos: coveredCount,
      totalVideos: total,
      percentage: total > 0 ? Math.round((coveredCount / total) * 100) : 0
    };
  }, [allVideos, coveredVideoIds]);

  const handleTogglePref = (language: string, type: 'manual' | 'automatic') => {
    setSelectedPrefs(prev => {
      const exists = prev.some(p => p.language === language && p.type === type);
      if (exists) {
        return prev.filter(p => !(p.language === language && p.type === type));
      } else {
        return [...prev, { language, type }];
      }
    });
  };

  const handleMovePref = (index: number, direction: 'up' | 'down') => {
    setSelectedPrefs(prev => {
      const newPrefs = [...prev];
      const targetIndex = direction === 'up' ? index - 1 : index + 1;
      if (targetIndex < 0 || targetIndex >= newPrefs.length) return prev;
      const temp = newPrefs[index];
      newPrefs[index] = newPrefs[targetIndex];
      newPrefs[targetIndex] = temp;
      return newPrefs;
    });
  };

  // Drag and drop handlers
  const handleDragStart = (e: React.DragEvent, index: number) => {
    setDraggedIndex(index);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  };

  const handleDrop = (e: React.DragEvent, targetIndex: number) => {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === targetIndex) return;

    setSelectedPrefs(prev => {
      const newPrefs = [...prev];
      const [movedItem] = newPrefs.splice(draggedIndex, 1);
      newPrefs.splice(targetIndex, 0, movedItem);
      return newPrefs;
    });
    setDraggedIndex(null);
  };

  const handleDragEnd = () => {
    setDraggedIndex(null);
  };

  const handleDownload = async (e: React.FormEvent) => {
    e.preventDefault();
    const list = inputs.split('\n').map(s => s.trim()).filter(Boolean);
    if (list.length === 0) {
      addError('Please enter your YouTube links first.');
      return;
    }

    if (selectedPrefs.length === 0) {
      addError('Please select at least one subtitle preference from the options list.');
      return;
    }

    setDownloading(true);
    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const payload = {
        inputs: list,
        format: format,
        preferences: selectedPrefs.map(p => ({
          language: p.language,
          type: p.type
        }))
      };

      const res = await fetch(`${apiUrl}/subtitle/download`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: controller.signal
      });

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || 'Download failed');
      }

      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'subtitles.zip';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err: any) {
      if (err.name !== 'AbortError' && !err.message?.includes('aborted')) {
        addError(err.message, { context: 'subtitle_download' });
      }
    } finally {
      setDownloading(false);
      abortControllerRef.current = null;
    }
  };

  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <div style={{ width: '100%', maxWidth: '1150px', margin: '0 auto', padding: '1.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        
        <div>
          <h1 style={{ fontSize: '1.8rem', fontWeight: 800, marginBottom: '0.5rem', letterSpacing: '-0.02em' }}>Subtitle Options & Multiset Selector</h1>
          <p style={{ color: theme.colors.textMuted, fontSize: '1.05rem' }}>
            Inspect available subtitle tracks across your videos/playlists, sort by remaining coverage impact, and manage your fallback preference order by dragging and dropping.
          </p>
        </div>

        {/* Step 1: Input Links & Get Options */}
        <form onSubmit={handleFetchOptions} style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', width: '100%' }}>
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem' }}>
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem' }}>1. Enter YouTube Links or IDs</h2>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem', marginBottom: '1rem' }}>
              Paste video or playlist links (one per line) to aggregate subtitle options:
            </p>
            <textarea
              value={inputs}
              onChange={(e) => setInputs(e.target.value)}
              placeholder="https://www.youtube.com/watch?v=...&#10;https://www.youtube.com/playlist?list=..."
              style={{
                width: '100%',
                minHeight: '150px',
                backgroundColor: theme.colors.bg,
                border: `1px solid ${theme.colors.borderColor}`,
                borderRadius: '6px',
                padding: '1.25rem',
                fontFamily: theme.fonts.mono,
                fontSize: '0.95rem',
                color: theme.colors.textMain,
                resize: 'vertical',
                lineHeight: '1.6'
              }}
            />

            <div style={{ marginTop: '1.5rem', display: 'flex', gap: '1rem', alignItems: 'center' }}>
              <button
                type="submit"
                disabled={loading}
                style={{
                  backgroundColor: theme.colors.primaryRed,
                  color: '#ffffff',
                  border: 'none',
                  borderRadius: '6px',
                  padding: '0.9rem 1.75rem',
                  fontWeight: 700,
                  fontSize: '1rem',
                  cursor: 'pointer',
                  opacity: loading ? 0.7 : 1,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  boxShadow: `0 4px 14px ${theme.colors.accentGlow}`
                }}
              >
                {loading && <div style={{ width: '1rem', height: '1rem', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', borderRadius: '50%', animation: 'spin 0.8s infinite linear' }}></div>}
                <span>Analyze Subtitle Options & Frequencies</span>
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

        {/* Step 2: Dedicated Selected Preferences Box (Draggable) */}
        {rawResponseData && (
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
              <div>
                <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.25rem' }}>2. Selected Subtitle Preferences Order (Drag & Drop)</h2>
                <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem' }}>
                  Total Videos: <strong style={{ color: theme.colors.textMain }}>{allVideos.length}</strong> | Combined Coverage: <strong style={{ color: '#4caf50' }}>{coverageMetrics.coveredVideos} / {coverageMetrics.totalVideos} ({coverageMetrics.percentage}%)</strong>
                </p>
              </div>
            </div>

            {selectedPrefs.length > 0 ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                <p style={{ fontSize: '0.8rem', color: theme.colors.textMuted, fontStyle: 'italic' }}>
                  💡 Drag and drop cards to reorder fallback priority, or use the move buttons.
                </p>
                {selectedPrefs.map((pref, idx) => (
                  <div
                    key={`${pref.type}_${pref.language}`}
                    draggable
                    onDragStart={(e) => handleDragStart(e, idx)}
                    onDragOver={(e) => handleDragOver(e, idx)}
                    onDrop={(e) => handleDrop(e, idx)}
                    onDragEnd={handleDragEnd}
                    style={{
                      backgroundColor: theme.colors.bg,
                      border: `1px solid ${theme.colors.primaryRed}`,
                      borderRadius: '6px',
                      padding: '0.75rem 1rem',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      cursor: 'grab',
                      opacity: draggedIndex === idx ? 0.5 : 1,
                      transition: 'background-color 0.2s'
                    }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                      <span style={{ cursor: 'grab', color: theme.colors.textMuted, fontWeight: 800 }}>⠿</span>
                      <span style={{
                        backgroundColor: theme.colors.primaryRed,
                        color: '#fff',
                        width: '1.5rem',
                        height: '1.5rem',
                        borderRadius: '50%',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: '0.8rem',
                        fontWeight: 700
                      }}>
                        {idx + 1}
                      </span>
                      <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>
                        {pref.language}
                      </span>
                      <span style={{
                        fontSize: '0.75rem',
                        fontWeight: 700,
                        textTransform: 'uppercase',
                        backgroundColor: pref.type === 'manual' ? '#2e7d32' : '#c62828',
                        color: '#fff',
                        padding: '0.1rem 0.4rem',
                        borderRadius: '4px'
                      }}>
                        {pref.type}
                      </span>
                    </div>

                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <button
                        type="button"
                        disabled={idx === 0}
                        onClick={() => handleMovePref(idx, 'up')}
                        style={{
                          backgroundColor: 'transparent',
                          color: idx === 0 ? theme.colors.borderColor : theme.colors.textMain,
                          border: `1px solid ${theme.colors.borderColor}`,
                          borderRadius: '4px',
                          padding: '0.2rem 0.5rem',
                          cursor: idx === 0 ? 'not-allowed' : 'pointer',
                          fontSize: '0.8rem'
                        }}
                      >
                        ↑ Up
                      </button>
                      <button
                        type="button"
                        disabled={idx === selectedPrefs.length - 1}
                        onClick={() => handleMovePref(idx, 'down')}
                        style={{
                          backgroundColor: 'transparent',
                          color: idx === selectedPrefs.length - 1 ? theme.colors.borderColor : theme.colors.textMain,
                          border: `1px solid ${theme.colors.borderColor}`,
                          borderRadius: '4px',
                          padding: '0.2rem 0.5rem',
                          cursor: idx === selectedPrefs.length - 1 ? 'not-allowed' : 'pointer',
                          fontSize: '0.8rem'
                        }}
                      >
                        ↓ Down
                      </button>
                      <button
                        type="button"
                        onClick={() => handleTogglePref(pref.language, pref.type)}
                        style={{
                          backgroundColor: 'transparent',
                          color: theme.colors.primaryRed,
                          border: `1px solid ${theme.colors.primaryRed}`,
                          borderRadius: '4px',
                          padding: '0.2rem 0.6rem',
                          cursor: 'pointer',
                          fontWeight: 700,
                          fontSize: '0.8rem'
                        }}
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem', fontStyle: 'italic' }}>
                No subtitle preferences selected yet. Click any available track below to add it to your multiset priority list.
              </p>
            )}
          </div>
        )}

        {/* Step 3: Available Subtitle Options Grid (with live remaining impact count) */}
        {rawResponseData && (
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
              <div>
                <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.25rem' }}>3. Available Subtitle Options (Live Coverage Impact)</h2>
                <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem' }}>
                  Click items to add or remove them. Remaining coverage updates live as you select preferences.
                </p>
              </div>
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search language (e.g. en, es)..."
                style={{
                  backgroundColor: theme.colors.bg,
                  border: `1px solid ${theme.colors.borderColor}`,
                  borderRadius: '6px',
                  padding: '0.6rem 1rem',
                  fontSize: '0.9rem',
                  color: theme.colors.textMain,
                  minWidth: '240px'
                }}
              />
            </div>

            {subtitleStats.length > 0 ? (
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
                gap: '1rem',
                maxHeight: '450px',
                overflowY: 'auto',
                paddingRight: '0.5rem'
              }}>
                {filteredStats.map((stat) => {
                  const isSelected = selectedPrefs.some(p => p.language === stat.languageCode && p.type === stat.type);
                  const percentage = Math.round((stat.count / stat.totalVideos) * 100);
                  const remainingPercentage = Math.round((stat.remainingCount / stat.totalVideos) * 100);

                  return (
                    <div
                      key={`${stat.type}_${stat.languageCode}`}
                      onClick={() => handleTogglePref(stat.languageCode, stat.type)}
                      style={{
                        backgroundColor: isSelected ? 'rgba(255, 30, 30, 0.15)' : theme.colors.bg,
                        border: `1px solid ${isSelected ? theme.colors.primaryRed : theme.colors.borderColor}`,
                        borderRadius: '6px',
                        padding: '1rem',
                        cursor: 'pointer',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '0.5rem',
                        transition: 'all 0.2s ease'
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <span style={{ fontWeight: 700, fontSize: '1rem', color: theme.colors.textMain }}>
                          {stat.languageName} ({stat.languageCode})
                        </span>
                        <span style={{
                          fontSize: '0.75rem',
                          fontWeight: 700,
                          textTransform: 'uppercase',
                          backgroundColor: stat.type === 'manual' ? '#2e7d32' : '#c62828',
                          color: '#fff',
                          padding: '0.15rem 0.5rem',
                          borderRadius: '4px'
                        }}>
                          {stat.type}
                        </span>
                      </div>

                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.85rem', color: theme.colors.textMuted }}>
                        <span>Total Track Coverage: {stat.count} / {stat.totalVideos} ({percentage}%)</span>
                      </div>

                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.85rem' }}>
                        <span style={{ color: theme.colors.textMuted }}>New Uncovered Videos:</span>
                        <strong style={{ color: stat.remainingCount > 0 ? '#4caf50' : '#888' }}>
                          +{stat.remainingCount} ({remainingPercentage}%)
                        </strong>
                      </div>

                      {/* Progress bar */}
                      <div style={{ width: '100%', height: '6px', backgroundColor: theme.colors.borderColor, borderRadius: '3px', overflow: 'hidden' }}>
                        <div style={{ width: `${remainingPercentage}%`, height: '100%', backgroundColor: isSelected ? theme.colors.primaryRed : '#4caf50' }}></div>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem' }}>No subtitle tracks found across the given inputs.</p>
            )}
          </div>
        )}

        {/* Step 4: Configure & Download Subtitles */}
        {rawResponseData && (
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem' }}>
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem' }}>4. Download Subtitles (ZIP Package)</h2>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem', marginBottom: '1.5rem' }}>
              Choose your format and trigger batch download of all matched subtitle files.
            </p>

            <form onSubmit={handleDownload} style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1.5rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.5rem' }}>
                    Output Format
                  </label>
                  <select
                    value={format}
                    onChange={(e) => setFormat(e.target.value)}
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
                    <option value="vtt">WebVTT (.vtt)</option>
                    <option value="srt">SubRip (.srt)</option>
                    <option value="json3">JSON3 (.json3)</option>
                    <option value="txt">Plain Text (.txt)</option>
                  </select>
                </div>
              </div>

              <div style={{ marginTop: '1rem', display: 'flex', gap: '1rem', alignItems: 'center' }}>
                <button
                  type="submit"
                  disabled={downloading}
                  style={{
                    backgroundColor: theme.colors.primaryRed,
                    color: '#ffffff',
                    border: 'none',
                    borderRadius: '6px',
                    padding: '1rem 2rem',
                    fontWeight: 700,
                    fontSize: '1rem',
                    cursor: 'pointer',
                    opacity: downloading ? 0.7 : 1,
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.75rem',
                    boxShadow: `0 4px 14px ${theme.colors.accentGlow}`
                  }}
                >
                  {downloading && <div style={{ width: '1rem', height: '1rem', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', borderRadius: '50%', animation: 'spin 0.8s infinite linear' }}></div>}
                  <span>Download Subtitles ZIP Package</span>
                </button>

                {downloading && (
                  <button
                    type="button"
                    onClick={handleCancel}
                    style={{
                      backgroundColor: 'transparent',
                      color: theme.colors.primaryRed,
                      border: `1px solid ${theme.colors.primaryRed}`,
                      borderRadius: '6px',
                      padding: '1rem 1.5rem',
                      fontWeight: 700,
                      fontSize: '1rem',
                      cursor: 'pointer'
                    }}
                  >
                    Cancel
                  </button>
                )}
              </div>
            </form>
          </div>
        )}

      </div>
    </div>
  );
}
