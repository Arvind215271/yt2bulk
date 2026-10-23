'use client';

import React, { useState } from 'react';
import Header from '@/components/Header';
import { availableFields } from '@/constants/fields';
import { theme } from '@/theme/theme';
import { useError } from '@/components/ErrorContext';

import { checkPreflight, PreflightResponse } from '@/services/preflight';
import PreflightResultCard from '@/components/PreflightResultCard';

export default function ExportPage() {
  const { addError } = useError();
  const [inputs, setInputs] = useState('');
  const [format, setFormat] = useState('json');
  const [playerClient, setPlayerClient] = useState('default');
  const [selectedFields, setSelectedFields] = useState<string[]>([
    'id', 'title', 'channel_title', 'duration', 'transcript_text', 'transcript_signal'
  ]);

  const [loading, setLoading] = useState(false);
  const [output, setOutput] = useState<string | null>(null);
  const [exportStats, setExportStats] = useState<any>(null);
  const [lastBlob, setLastBlob] = useState<Blob | null>(null);
  const [filename, setFilename] = useState('export.json');
  const [preflightData, setPreflightData] = useState<PreflightResponse | null>(null);
  const [preflightLoading, setPreflightLoading] = useState(false);

  const abortControllerRef = React.useRef<AbortController | null>(null);

  const handleCancel = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setLoading(false);
    setSubtitleOptionsLoading(false);
    setOutput('Request cancelled by user.');
  };

  // Subtitle options integration using /subtitle/options endpoint
  const [subtitleOptionsLoading, setSubtitleOptionsLoading] = useState(false);
  const [rawSubtitleOptions, setRawSubtitleOptions] = useState<any>(null);
  const [subtitleSearchQuery, setSubtitleSearchQuery] = useState('');
  const [selectedSubtitlePrefs, setSelectedSubtitlePrefs] = useState<Array<{ language: string; type: 'manual' | 'automatic' }>>([]);

  const handlePreflight = async () => {
    const list = inputs.split('\n').map((s: string) => s.trim()).filter(Boolean);
    if (list.length === 0) {
      alert('Please enter at least one YouTube link or ID first.');
      return;
    }

    setPreflightLoading(true);
    setPreflightData(null);
    setOutput(null);

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const data = await checkPreflight(list, playerClient, controller.signal);
      setPreflightData(data);
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        addError(err.message, { context: 'preflight', details: err });
      }
    } finally {
      setPreflightLoading(false);
      abortControllerRef.current = null;
    }
  };

  const handleFetchSubtitleOptions = async () => {
    const list = inputs.split('\n').map((s: string) => s.trim()).filter(Boolean);
    if (list.length === 0) {
      alert('Please enter at least one YouTube link or ID first.');
      return;
    }

    setSubtitleOptionsLoading(true);
    setRawSubtitleOptions(null);
    setSelectedSubtitlePrefs([]);

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      console.log('Sending subtitle options request from export page to:', `${apiUrl}/subtitle/options`, { inputs: list });
      const res = await fetch(`${apiUrl}/subtitle/options`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ inputs: list }),
        signal: controller.signal
      });

      const text = await res.text();
      if (!res.ok) {
        console.error('Failed to fetch subtitle options from export page, status:', res.status, text);
        throw new Error(text || 'Failed to fetch subtitle options');
      }

      const parsed = JSON.parse(text);
      console.log('Successfully fetched subtitle options from export page:', parsed);
      setRawSubtitleOptions(parsed);
    } catch (err: any) {
      if (err.name !== 'AbortError' && !err.message?.includes('aborted')) {
        addError(err.message, { context: 'subtitle_options' });
      }
    } finally {
      setSubtitleOptionsLoading(false);
      abortControllerRef.current = null;
    }
  };

  const allSubtitleVideos = React.useMemo(() => {
    if (!rawSubtitleOptions) return [];
    const vids: any[] = [];
    const items = Array.isArray(rawSubtitleOptions) ? rawSubtitleOptions : (rawSubtitleOptions.resources || [rawSubtitleOptions]);

    for (const item of items) {
      if (!item) continue;
      if (item.type === 'video' && item.video) {
        vids.push(item.video);
      } else if (item.subtitle_metadata) {
        vids.push(item);
      } else if (item.type === 'playlist' && item.playlist && item.playlist.videos) {
        vids.push(...item.playlist.videos);
      } else if (item.videos) {
        vids.push(...item.videos);
      }
    }
    return vids;
  }, [rawSubtitleOptions]);

  const subtitleStats = React.useMemo(() => {
    const totalVideos = allSubtitleVideos.length;
    if (totalVideos === 0) return [];

    const statsMap = new Map<string, { languageCode: string; languageName: string; type: 'manual' | 'automatic'; count: number; totalVideos: number }>();

    for (const vid of allSubtitleVideos) {
      const manualTracks = vid.subtitle_metadata?.manual || [];
      const autoTracks = vid.subtitle_metadata?.automatic || [];

      for (const track of manualTracks) {
        if (!track.languageCode) continue;
        const key = `manual_${track.languageCode}`;
        if (!statsMap.has(key)) {
          statsMap.set(key, {
            languageCode: track.languageCode,
            languageName: track.languageName || track.languageCode,
            type: 'manual',
            count: 0,
            totalVideos
          });
        }
        statsMap.get(key)!.count++;
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
            totalVideos
          });
        }
        statsMap.get(key)!.count++;
      }
    }

    const list = Array.from(statsMap.values());
    list.sort((a, b) => b.count - a.count);
    return list;
  }, [allSubtitleVideos]);

  const filteredSubtitleStats = React.useMemo(() => {
    if (!subtitleSearchQuery.trim()) return subtitleStats;
    const q = subtitleSearchQuery.toLowerCase();
    return subtitleStats.filter(s =>
      s.languageCode.toLowerCase().includes(q) ||
      s.languageName.toLowerCase().includes(q) ||
      s.type.toLowerCase().includes(q)
    );
  }, [subtitleStats, subtitleSearchQuery]);

  const handleToggleSubtitlePref = (language: string, type: 'manual' | 'automatic') => {
    setSelectedSubtitlePrefs(prev => {
      const exists = prev.some(p => p.language === language && p.type === type);
      if (exists) {
        return prev.filter(p => !(p.language === language && p.type === type));
      } else {
        return [...prev, { language, type }];
      }
    });
  };

  const handleFieldToggle = (field: string) => {
    setSelectedFields(prev =>
      prev.includes(field) ? prev.filter(f => f !== field) : [...prev, field]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const list = inputs.split('\n').map((s: string) => s.trim()).filter(Boolean);
    if (list.length === 0) {
      alert('Please enter at least one YouTube link or ID.');
      return;
    }

    setLoading(true);
    setOutput('Processing your export...');

    const includeTranscript = selectedFields.includes('transcript_text') || selectedFields.includes('subtitle_metadata');
    const includeSignal = selectedFields.includes('transcript_signal');

    const payload = {
      inputs: list,
      format: format,
      player_client: playerClient,
      fields: selectedFields,
      transcript: includeTranscript ? { output: 'plain-text' } : null,
      signal: includeSignal ? {} : null,
      preferences: selectedSubtitlePrefs.length > 0 ? selectedSubtitlePrefs : undefined
    };

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const res = await fetch(`${apiUrl}/export`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: controller.signal
      });

      if (!res.ok) {
        const errText = await res.text();
        let errorMessage = errText || 'Export failed';
        let errorData = undefined;
        try {
          const errorJson = JSON.parse(errText);
          if (errorJson && errorJson.message) {
            errorMessage = errorJson.message;
          }
          if (errorJson && errorJson.data) {
            errorData = errorJson.data;
          }
        } catch (e) {
          console.warn('[ExportPage] Failed to parse error response as JSON', e);
        }
        const error = new Error(errorMessage);
        (error as any).data = errorData;
        throw error;
      }

      const contentType = res.headers.get('content-type') || '';
      const blob = await res.blob();
      setLastBlob(blob);

      let ext = 'json';
      if (format === 'csv') ext = 'csv';
      else if (format === 'markdown') ext = 'md';
      else if (format === 'yaml') ext = 'yaml';
      else if (format === 'xml') ext = 'xml';
      else if (format === 'excel') ext = 'xlsx';

      setFilename(`export_${Date.now()}.${ext}`);

      const text = await blob.text();
      if (contentType.includes('json') || format === 'json' || format === 'yaml' || format === 'xml') {
        try {
          const parsed = JSON.parse(text);
          if (parsed.statistics) {
            setExportStats(parsed.statistics);
            const { statistics, ...cleaned } = parsed;
            setOutput(JSON.stringify(cleaned, null, 2));
          } else {
            setExportStats(null);
            setOutput(JSON.stringify(parsed, null, 2));
          }
        } catch {
          setExportStats(null);
          setOutput(text);
        }
      } else {
        setOutput(text.slice(0, 5000) + (text.length > 5000 ? '\n\n[Output truncated for preview]' : ''));
      }
    } catch (err: any) {
      if (err.name !== 'AbortError' && !err.message?.includes('aborted')) {
        addError(err.message, err.data || { context: 'export' });
        setOutput('Export failed.');
      }
    } finally {
      setLoading(false);
      abortControllerRef.current = null;
    }
  };

  const handleCopy = () => {
    if (!output) return;
    navigator.clipboard.writeText(output).then(() => {
      alert('Copied to clipboard!');
    }).catch(err => {
      console.error('Failed to copy: ', err);
    });
  };

  const handleDownload = () => {
    if (!lastBlob) return;
    const url = URL.createObjectURL(lastBlob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <div style={{ width: '100%', maxWidth: '1100px', margin: '0 auto', padding: '1.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 800, marginBottom: '0.5rem', letterSpacing: '-0.02em' }}>Export YouTube Data</h1>
          <p style={{ color: theme.colors.textMuted, fontSize: '0.95rem' }}>
            Paste your YouTube video or playlist links below, choose the format and fields you want, and instantly export everything.
          </p>
        </div>

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', width: '100%' }}>
          
          {/* Big Wide Central Box for Links */}
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.5rem' }}>
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.5rem' }}>1. Paste your YouTube Links or IDs</h2>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.85rem', marginBottom: '1rem' }}>
              Put as many links as you want here (videos, playlists, shorts), with each link on its own separate line:
            </p>
            <textarea
              value={inputs}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setInputs(e.target.value)}
              placeholder="https://www.youtube.com/watch?v=...&#10;https://www.youtube.com/playlist?list=...&#10;https://youtu.be/..."
              style={{
                width: '100%',
                minHeight: '150px',
                backgroundColor: theme.colors.bg,
                border: `1px solid ${theme.colors.borderColor}`,
                borderRadius: '6px',
                padding: '1rem',
                fontFamily: theme.fonts.mono,
                fontSize: '0.9rem',
                color: theme.colors.textMain,
                resize: 'vertical',
                lineHeight: '1.5'
              }}
            />

            <div style={{ marginTop: '1rem', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 600, color: theme.colors.textMuted }}>
                Choose File Format:
              </label>
              <select
                value={format}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setFormat(e.target.value)}
                style={{
                  backgroundColor: theme.colors.bg,
                  border: `1px solid ${theme.colors.borderColor}`,
                  borderRadius: '6px',
                  padding: '0.5rem',
                  fontSize: '0.9rem',
                  color: theme.colors.textMain,
                  width: '100%'
                }}
              >
                <option value="json">JSON (.json)</option>
                <option value="csv">CSV (.csv)</option>
                <option value="markdown">Markdown (.md)</option>
                <option value="yaml">YAML (.yaml)</option>
                <option value="xml">XML (.xml)</option>
                <option value="excel">Excel (.xlsx)</option>
              </select>
            </div>
            <div style={{ marginTop: '1rem', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 600, color: theme.colors.textMuted }}>
                Player Client:
              </label>
              <select
                value={playerClient}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setPlayerClient(e.target.value)}
                style={{
                  backgroundColor: theme.colors.bg,
                  border: `1px solid ${theme.colors.borderColor}`,
                  borderRadius: '6px',
                  padding: '0.5rem',
                  fontSize: '0.9rem',
                  color: theme.colors.textMain,
                  width: '100%'
                }}
              >
                <option value="default">Default</option>
                <option value="web">Web</option>
                <option value="web_safari">Web Safari</option>
                <option value="web_embedded">Web Embedded</option>
                <option value="mweb">Mobile Web (mweb)</option>
                <option value="ios">iOS</option>
                <option value="android">Android</option>
                <option value="android_vr">Android VR</option>
                <option value="android_sdkless">Android (SDK-less)</option>
                <option value="tv">TV</option>
                <option value="tv_simply">TV Simply</option>
                <option value="tv_embedded">TV Embedded</option>
                <option value="visionos">VisionOS</option>
                <option value="tv_samsung">TV Samsung</option>
                <option value="all">All</option>
              </select>
            </div>
          </div>

          {/* Fields Selection Box Below */}
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '2rem' }}>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '0.5rem' }}>2. Select Information to Extract</h2>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem', marginBottom: '1.5rem' }}>
              Check what details you want included in your export. Selecting transcripts or word stats will automatically fetch them for you.
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              <div>
                <h3 style={{ fontSize: '0.9rem', color: theme.colors.primaryRed, textTransform: 'uppercase', fontWeight: 700, marginBottom: '0.75rem' }}>
                  Basic Info & Stats
                </h3>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '0.75rem' }}>
                  {availableFields.filter(f => !f.includes('description') && !f.includes('transcript') && !f.includes('subtitle')).map(field => (
                    <label key={field} style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', fontSize: '0.9rem', color: theme.colors.textMain, cursor: 'pointer' }}>
                      <input
                        type="checkbox"
                        checked={selectedFields.includes(field)}
                        onChange={() => handleFieldToggle(field)}
                        style={{ accentColor: theme.colors.primaryRed, width: '1.1rem', height: '1.1rem' }}
                      />
                      {field}
                    </label>
                  ))}
                </div>
              </div>

              <div>
                <h3 style={{ fontSize: '0.9rem', color: theme.colors.primaryRed, textTransform: 'uppercase', fontWeight: 700, marginBottom: '0.75rem' }}>
                  Description Details
                </h3>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '0.75rem' }}>
                  {availableFields.filter(f => f.includes('description')).map(field => (
                    <label key={field} style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', fontSize: '0.9rem', color: theme.colors.textMain, cursor: 'pointer' }}>
                      <input
                        type="checkbox"
                        checked={selectedFields.includes(field)}
                        onChange={() => handleFieldToggle(field)}
                        style={{ accentColor: theme.colors.primaryRed, width: '1.1rem', height: '1.1rem' }}
                      />
                      {field}
                    </label>
                  ))}
                </div>
              </div>

              <div>
                <h3 style={{ fontSize: '0.9rem', color: theme.colors.primaryRed, textTransform: 'uppercase', fontWeight: 700, marginBottom: '0.75rem' }}>
                  Transcript & Word Signals
                </h3>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '0.75rem' }}>
                  {availableFields.filter(f => f.includes('transcript') || f.includes('subtitle')).map(field => (
                    <label key={field} style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', fontSize: '0.9rem', color: theme.colors.textMain, cursor: 'pointer' }}>
                      <input
                        type="checkbox"
                        checked={selectedFields.includes(field)}
                        onChange={() => handleFieldToggle(field)}
                        style={{ accentColor: theme.colors.primaryRed, width: '1.1rem', height: '1.1rem' }}
                      />
                      {field}
                    </label>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Subtitle Options & Multiset Selector Box */}
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '2rem', display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-wrap', gap: '1rem' }}>
              <div>
                <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '0.25rem' }}>3. Subtitle Options & Preferences</h2>
                <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem' }}>
                  Analyze available subtitle tracks for your inputs and select specific languages/types to fetch.
                </p>
              </div>
              <button
                type="button"
                onClick={handleFetchSubtitleOptions}
                disabled={subtitleOptionsLoading}
                style={{
                  backgroundColor: theme.colors.bg,
                  color: theme.colors.textMain,
                  border: `1px solid ${theme.colors.borderColor}`,
                  borderRadius: '6px',
                  padding: '0.6rem 1.25rem',
                  fontWeight: 600,
                  fontSize: '0.9rem',
                  cursor: 'pointer',
                  opacity: subtitleOptionsLoading ? 0.7 : 1,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem'
                }}
              >
                {subtitleOptionsLoading && <div style={{ width: '0.9rem', height: '0.9rem', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', borderRadius: '50%', animation: 'spin 0.8s infinite linear' }}></div>}
                <span>Fetch Available Subtitle Options</span>
              </button>
            </div>

            {rawSubtitleOptions && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', paddingTop: '0.5rem', borderTop: `1px solid ${theme.colors.borderColor}` }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                  <span style={{ fontSize: '0.9rem', color: theme.colors.textMuted }}>
                    Total Videos Analyzed: <strong style={{ color: theme.colors.textMain }}>{allSubtitleVideos.length}</strong>
                  </span>
                  <input
                    type="text"
                    value={subtitleSearchQuery}
                    onChange={(e) => setSubtitleSearchQuery(e.target.value)}
                    placeholder="Search language (e.g. en, es)..."
                    style={{
                      backgroundColor: theme.colors.bg,
                      border: `1px solid ${theme.colors.borderColor}`,
                      borderRadius: '6px',
                      padding: '0.5rem 0.8rem',
                      fontSize: '0.85rem',
                      color: theme.colors.textMain,
                      minWidth: '220px'
                    }}
                  />
                </div>

                {filteredSubtitleStats.length > 0 ? (
                  <div style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
                    gap: '0.75rem',
                    maxHeight: '280px',
                    overflowY: 'auto',
                    paddingRight: '0.5rem'
                  }}>
                    {filteredSubtitleStats.map((stat) => {
                      const isSelected = selectedSubtitlePrefs.some(p => p.language === stat.languageCode && p.type === stat.type);
                      const percentage = Math.round((stat.count / stat.totalVideos) * 100);

                      return (
                        <div
                          key={`${stat.type}_${stat.languageCode}`}
                          onClick={() => handleToggleSubtitlePref(stat.languageCode, stat.type)}
                          style={{
                            backgroundColor: isSelected ? 'rgba(255, 30, 30, 0.15)' : theme.colors.bg,
                            border: `1px solid ${isSelected ? theme.colors.primaryRed : theme.colors.borderColor}`,
                            borderRadius: '6px',
                            padding: '0.8rem',
                            cursor: 'pointer',
                            display: 'flex',
                            flexDirection: 'column',
                            gap: '0.4rem',
                            transition: 'all 0.2s ease'
                          }}
                        >
                          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                            <span style={{ fontWeight: 700, fontSize: '0.9rem', color: theme.colors.textMain }}>
                              {stat.languageName} ({stat.languageCode})
                            </span>
                            <span style={{
                              fontSize: '0.7rem',
                              fontWeight: 700,
                              textTransform: 'uppercase',
                              backgroundColor: stat.type === 'manual' ? '#2e7d32' : '#c62828',
                              color: '#fff',
                              padding: '0.1rem 0.4rem',
                              borderRadius: '4px'
                            }}>
                              {stat.type}
                            </span>
                          </div>

                          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.8rem', color: theme.colors.textMuted }}>
                            <span>Covers: {stat.count}/{stat.totalVideos} videos</span>
                            <strong style={{ color: percentage > 50 ? '#4caf50' : theme.colors.textMain }}>{percentage}%</strong>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem', fontStyle: 'italic' }}>No subtitle tracks found or none fetched yet.</p>
                )}

                {/* Selected Preferences Summary Badge */}
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', alignItems: 'center', paddingTop: '0.5rem' }}>
                  <span style={{ fontSize: '0.85rem', fontWeight: 600, color: theme.colors.textMuted }}>Selected Preferences:</span>
                  {selectedSubtitlePrefs.length > 0 ? (
                    selectedSubtitlePrefs.map((p, idx) => (
                      <span key={idx} style={{
                        backgroundColor: theme.colors.primaryRed,
                        color: '#fff',
                        padding: '0.2rem 0.6rem',
                        borderRadius: '4px',
                        fontSize: '0.8rem',
                        fontWeight: 600,
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.4rem'
                      }}>
                        {p.language} ({p.type})
                        <span onClick={(e) => { e.stopPropagation(); handleToggleSubtitlePref(p.language, p.type); }} style={{ cursor: 'pointer', fontWeight: 800 }}>×</span>
                      </span>
                    ))
                  ) : (
                    <span style={{ fontSize: '0.85rem', color: theme.colors.textMuted, fontStyle: 'italic' }}>None selected (will use defaults if subtitle_metadata requested).</span>
                  )}
                </div>
              </div>
            )}
          </div>

          <div style={{ marginTop: '1rem', display: 'flex', gap: '1rem' }}>
            <button
              type="submit"
              disabled={loading}
              style={{
                flex: 1,
                backgroundColor: theme.colors.primaryRed,
                color: '#ffffff',
                border: 'none',
                borderRadius: '6px',
                padding: '1.1rem',
                fontWeight: 700,
                fontSize: '1.05rem',
                cursor: 'pointer',
                opacity: loading ? 0.7 : 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.75rem',
                boxShadow: `0 4px 14px ${theme.colors.accentGlow}`
              }}
            >
              {loading && <div style={{ width: '1.2rem', height: '1.2rem', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', borderRadius: '50%', animation: 'spin 0.8s infinite linear' }}></div>}
              <span>Start Export Now</span>
            </button>

            {(loading || subtitleOptionsLoading) && (
              <button
                type="button"
                onClick={handleCancel}
                style={{
                  backgroundColor: 'transparent',
                  color: theme.colors.textMain,
                  border: `1px solid ${theme.colors.borderColor}`,
                  borderRadius: '6px',
                  padding: '1.1rem 1.5rem',
                  fontWeight: 600,
                  fontSize: '1.05rem',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '0.5rem'
                }}
              >
                Cancel Export
              </button>
            )}
          </div>
        </form>

        {output && (
          <div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
            
            {/* Export Result Card */}
            <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '2rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                <h2 style={{ fontSize: '1.2rem', fontWeight: 700 }}>Your Export Result</h2>
                {output && (
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <button
                      onClick={handleCopy}
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
                      Copy to Clipboard
                    </button>
                    {lastBlob && (
                      <button
                        onClick={handleDownload}
                        style={{
                          backgroundColor: theme.colors.primaryRed,
                          color: '#ffffff',
                          border: 'none',
                          borderRadius: '6px',
                          padding: '0.6rem 1.2rem',
                          fontSize: '0.85rem',
                          fontWeight: 600,
                          cursor: 'pointer'
                        }}
                      >
                        Download File
                      </button>
                    )}
                  </div>
                )}
              </div>
              <pre style={{
                backgroundColor: theme.colors.bg,
                border: `1px solid ${theme.colors.borderColor}`,
                padding: '1.25rem',
                borderRadius: '6px',
                fontFamily: theme.fonts.mono,
                fontSize: '0.85rem',
                color: theme.colors.outputCode,
                overflowX: 'auto',
                maxHeight: '450px',
                lineHeight: '1.5'
              }}>
                {output}
              </pre>
            </div>

            {/* Processing Statistics Card */}
            {exportStats && (
              <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '2rem' }}>
                <h2 style={{ fontSize: '1.3rem', fontWeight: 700, marginBottom: '1.5rem', borderBottom: `1px solid ${theme.colors.borderColor}`, paddingBottom: '0.75rem' }}>
                  Processing Statistics Breakdown
                </h2>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                  
                  {/* 1. Resource Stats Table */}
                  <div style={{ backgroundColor: theme.colors.bg, borderRadius: '8px', border: `1px solid ${theme.colors.borderColor}`, overflow: 'hidden' }}>
                    <div style={{ backgroundColor: 'rgba(255,255,255,0.03)', padding: '0.75rem 1rem', borderBottom: `1px solid ${theme.colors.borderColor}`, fontWeight: 700, fontSize: '0.95rem', color: theme.colors.primaryRed }}>
                      Resource Statistics
                    </div>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', padding: '1rem', gap: '1rem' }}>
                      <div>
                        <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Requested</div>
                        <div style={{ fontSize: '1.2rem', fontWeight: 700 }}>{exportStats.ResourcesRequested ?? 0}</div>
                      </div>
                      <div>
                        <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Succeeded</div>
                        <div style={{ fontSize: '1.2rem', fontWeight: 700, color: '#4caf50' }}>{exportStats.ResourcesSucceeded ?? 0}</div>
                      </div>
                      <div>
                        <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Failures</div>
                        <div style={{ fontSize: '1.2rem', fontWeight: 700, color: '#c62828' }}>{exportStats.ResourceFailures ?? 0}</div>
                      </div>
                      <div>
                        <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Videos Processed</div>
                        <div style={{ fontSize: '1.2rem', fontWeight: 700 }}>{exportStats.VideosProcessed ?? 0}</div>
                      </div>
                    </div>
                  </div>

                  {/* 2. Metadata Stats Table */}
                  {exportStats.Metadata && (
                    <div style={{ backgroundColor: theme.colors.bg, borderRadius: '8px', border: `1px solid ${theme.colors.borderColor}`, overflow: 'hidden' }}>
                      <div style={{ backgroundColor: 'rgba(255,255,255,0.03)', padding: '0.75rem 1rem', borderBottom: `1px solid ${theme.colors.borderColor}`, fontWeight: 700, fontSize: '0.95rem', color: theme.colors.primaryRed }}>
                        Metadata Statistics
                      </div>
                      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))', padding: '1rem', gap: '1rem' }}>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Attempts</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Metadata.Attempts ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Successes</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700, color: '#4caf50' }}>{exportStats.Metadata.Successes ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Failures</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700, color: '#c62828' }}>{exportStats.Metadata.Failures ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Cache Hits</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Metadata.CacheHits ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Cache Misses</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Metadata.CacheMisses ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Upstream Fetches</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Metadata.UpstreamFetches ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>YT-DLP Requests</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Metadata.YTDLPRequests ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Fallbacks</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Metadata.Fallbacks ?? 0}</div>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* 3. Subtitle Stats Table */}
                  {exportStats.Subtitle && (
                    <div style={{ backgroundColor: theme.colors.bg, borderRadius: '8px', border: `1px solid ${theme.colors.borderColor}`, overflow: 'hidden' }}>
                      <div style={{ backgroundColor: 'rgba(255,255,255,0.03)', padding: '0.75rem 1rem', borderBottom: `1px solid ${theme.colors.borderColor}`, fontWeight: 700, fontSize: '0.95rem', color: theme.colors.primaryRed }}>
                        Subtitle Statistics
                      </div>
                      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))', padding: '1rem', gap: '1rem' }}>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Attempts</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Subtitle.Attempts ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Successes</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700, color: '#4caf50' }}>{exportStats.Subtitle.Successes ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Failures</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700, color: '#c62828' }}>{exportStats.Subtitle.Failures ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Manual Requests</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Subtitle.ManualRequests ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Automatic Requests</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Subtitle.AutomaticRequests ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Cache Hits</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Subtitle.CacheHits ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Upstream Fetches</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Subtitle.UpstreamFetches ?? 0}</div>
                        </div>
                        <div>
                          <div style={{ fontSize: '0.8rem', color: theme.colors.textMuted }}>Fallbacks</div>
                          <div style={{ fontSize: '1.1rem', fontWeight: 700 }}>{exportStats.Subtitle.Fallbacks ?? 0}</div>
                        </div>
                      </div>
                    </div>
                  )}

                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
