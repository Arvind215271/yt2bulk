'use client';

import React, { useState } from 'react';
import Header from '@/components/Header';
import { theme } from '@/theme/theme';
import { useError } from '@/components/ErrorContext';

export default function TranscriptPage() {
  const { addError } = useError();
  const [inputs, setInputs] = useState('');
  const [analysisType, setAnalysisType] = useState<'word-stats' | 'window-stats'>('window-stats');
  const [useHeavy, setUseHeavy] = useState(true);
  const [loading, setLoading] = useState(false);
  const [output, setOutput] = useState<any>(null);

  const abortControllerRef = React.useRef<AbortController | null>(null);

  const handleCancel = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setLoading(false);
    addError('Transcript request cancelled by user.');
  };

  const handleFetchTranscript = async (e: React.FormEvent) => {
    e.preventDefault();
    const list = inputs.split('\n').map(s => s.trim()).filter(Boolean);
    if (list.length === 0) {
      addError('Please enter at least one YouTube link or ID.');
      return;
    }

    setLoading(true);
    setOutput(null);

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const res = await fetch(`${apiUrl}/transcript`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          inputs: list,
          signal: {
            analysis: analysisType,
            use_heavy_stopwords: useHeavy,
            min_freq: 3,
            depth: 1,
            window_size: 300
          }
        }),
        signal: controller.signal
      });

      const text = await res.text();
      if (!res.ok) {
        throw new Error(text || 'Failed to fetch transcript');
      }

      const parsed = JSON.parse(text);
      setOutput(parsed);
    } catch (err: any) {
      if (err.name === 'AbortError' || err.message?.includes('aborted')) {
        addError('Transcript request cancelled by user.');
      } else {
        addError(err.message);
      }
    } finally {
      setLoading(false);
      abortControllerRef.current = null;
    }
  };

  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <div style={{ width: '100%', maxWidth: '1100px', margin: '0 auto', padding: '1.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        
        <div>
          <h1 style={{ fontSize: '1.8rem', fontWeight: 800, marginBottom: '0.5rem', letterSpacing: '-0.02em' }}>Transcript Viewer <span style={{fontSize: '0.9rem', color: theme.colors.primaryRed, fontWeight: 500, verticalAlign: 'middle', marginLeft: '0.5rem'}}>(Local Instance)</span></h1>
          <p style={{ color: theme.colors.textMuted, fontSize: '1.05rem' }}>
            Fetch raw JSON3/structured transcripts directly from YouTube videos using high-performance internal pipelines.
          </p>
        </div>

        <form onSubmit={handleFetchTranscript} style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', width: '100%' }}>
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem' }}>
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.75rem' }}>Enter YouTube Links or IDs</h2>
            <p style={{ color: theme.colors.textMuted, fontSize: '0.9rem', marginBottom: '1rem' }}>
              Paste one or more video links or IDs (one per line):
            </p>
            <textarea
              value={inputs}
              onChange={(e) => setInputs(e.target.value)}
              placeholder="https://www.youtube.com/watch?v=..."
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
                lineHeight: '1.6',
                marginBottom: '1.0rem'
              }}
            />
            <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem', color: theme.colors.textMuted }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                Analysis:
                <select value={analysisType} onChange={(e) => setAnalysisType(e.target.value as any)} style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, border: `1px solid ${theme.colors.borderColor}`, padding: '0.25rem', borderRadius: '4px' }}>
                  <option value="word-stats">Word Stats</option>
                  <option value="window-stats">Window Stats</option>
                </select>
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input type="checkbox" checked={useHeavy} onChange={(e) => setUseHeavy(e.target.checked)} />
                Use Heavy Stopwords
              </label>
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
                  padding: '0.9rem 2rem',
                  fontWeight: 700,
                  fontSize: '1rem',
                  cursor: 'pointer',
                  opacity: loading ? 0.7 : 1,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.75rem',
                  boxShadow: `0 4px 14px ${theme.colors.accentGlow}`
                }}
              >
                {loading && <div style={{ width: '1rem', height: '1rem', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', borderRadius: '50%', animation: 'spin 0.8s infinite linear' }}></div>}
                <span>Fetch Transcripts</span>
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

        {output && (
          <div style={{ backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <h2 style={{ fontSize: '1.1rem', fontWeight: 700 }}>Transcript Result</h2>
            <pre style={{
              backgroundColor: theme.colors.bg,
              border: `1px solid ${theme.colors.borderColor}`,
              padding: '1.25rem',
              borderRadius: '6px',
              fontFamily: theme.fonts.mono,
              fontSize: '0.85rem',
              color: theme.colors.outputCode,
              overflowX: 'auto',
              maxHeight: '500px',
              lineHeight: '1.5'
            }}>
              {JSON.stringify(output, null, 2)}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
