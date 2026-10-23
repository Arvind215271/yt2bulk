'use client';

import React, { useState, useRef } from 'react';
import Header from '@/components/Header';
import { theme } from '@/theme/theme';
import { checkPreflight, PreflightResponse } from '@/services/preflight';
import PreflightResultCard from '@/components/PreflightResultCard';
import { useError } from '@/components/ErrorContext';

export default function PreflightPage() {
  const { addError } = useError();
  const [inputs, setInputs] = useState('');
  const [playerClient, setPlayerClient] = useState('default');
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<PreflightResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const handlePreflight = async (e: React.FormEvent) => {
    e.preventDefault();
    const list = inputs.split('\n').map((s: string) => s.trim()).filter(Boolean);
    if (list.length === 0) {
      const msg = 'Please enter at least one YouTube link or ID.';
      setError(msg);
      addError(msg);
      return;
    }

    setLoading(true);
    setError(null);
    setData(null);

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const result = await checkPreflight(list, playerClient, controller.signal);
      setData(result);
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        const msg = err.message || 'An unexpected error occurred.';
        setError(msg);
        const details = err.details?.data || err.details;
        addError(msg, details);
      }
    } finally {
      setLoading(false);
      abortControllerRef.current = null;
    }
  };

  const handleProceed = () => {
    // In a real application, this would likely navigate to the export page 
    // or initiate the next step in the pipeline.
    alert('Proceeding to next step...');
  };

  return (
    <div style={{ backgroundColor: theme.colors.bg, color: theme.colors.textMain, minHeight: '100vh', display: 'flex', flexDirection: 'column', fontFamily: theme.fonts.sans, width: '100%', margin: 0 }}>
      <Header />
      <div style={{ width: '100%', maxWidth: '900px', margin: '0 auto', padding: '1.5rem 1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        
        <div>
          <h1 style={{ fontSize: '1.8rem', fontWeight: 800, marginBottom: '0.75rem', letterSpacing: '-0.02em' }}>Preflight Check</h1>
          <p style={{ color: theme.colors.textMuted, fontSize: '1.1rem', lineHeight: '1.6' }}>
            Verify your YouTube links before processing to ensure they are valid and accessible.
          </p>
        </div>
        
        <form onSubmit={handlePreflight} style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', backgroundColor: theme.colors.panelBg, border: `1px solid ${theme.colors.borderColor}`, borderRadius: theme.spacing.borderRadius, padding: '1.25rem' }}>
          <textarea
            value={inputs}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setInputs(e.target.value)}
            placeholder="https://www.youtube.com/watch?v=...&#10;https://www.youtube.com/playlist?list=..."
            style={{
              width: '100%',
              minHeight: '200px',
              backgroundColor: theme.colors.bg,
              border: `1px solid ${theme.colors.borderColor}`,
              borderRadius: '6px',
              padding: '1rem',
              fontFamily: theme.fonts.mono,
              fontSize: '0.95rem',
              color: theme.colors.textMain,
              resize: 'vertical'
            }}
          />
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
              <option value="android">Android</option>
              <option value="android_vr">Android VR</option>
              <option value="visionos">VisionOS</option>
              <option value="tv_samsung">TV Samsung</option>
              <option value="all">All</option>
            </select>
          </div>
          <button
            type="submit"
            disabled={loading}
            style={{
              backgroundColor: theme.colors.primaryRed,
              color: '#ffffff',
              border: 'none',
              borderRadius: '6px',
              padding: '0.8rem 1.5rem',
              fontWeight: 600,
              cursor: loading ? 'not-allowed' : 'pointer',
              opacity: loading ? 0.7 : 1
            }}
          >
            {loading ? 'Analyzing...' : 'Run Preflight'}
          </button>
        </form>

        {error && (
          <div style={{ color: theme.colors.primaryRed, backgroundColor: 'rgba(255, 30, 30, 0.1)', padding: '1rem', borderRadius: '6px', border: `1px solid ${theme.colors.primaryRed}` }}>
            {error}
          </div>
        )}

        <PreflightResultCard
          data={data}
          loading={loading}
          onProceed={handleProceed}
          proceedLabel="Continue to Export"
        />

      </div>
    </div>
  );
}
