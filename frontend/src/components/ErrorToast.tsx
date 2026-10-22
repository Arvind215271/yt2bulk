'use client';

import React from 'react';
import ErrorAlert from './ErrorAlert';

interface ErrorToastProps {
  message: string;
  data?: any;
  onDismiss?: () => void;
  autoDismissMs?: number;
}

import { parseAppError } from '@/utils/errorParser';

export default function ErrorToast({ message, data, onDismiss, autoDismissMs }: ErrorToastProps) {
  console.log('[ErrorToast] Rendering', { message, data });

  const parsed = parseAppError({ message, data });
  const displayMessage = parsed.message;
  const displayFields = parsed.fields;
  const displayData = parsed.data;

  let detailContent = null;
  if (displayFields && displayFields.length > 0) {
    detailContent = (
      <div style={{ marginTop: '0.4rem', display: 'flex', flexDirection: 'column', gap: '0.2rem', fontSize: '0.8rem', color: '#cbd5e1' }}>
        {displayFields.map((item: any, idx: number) => (
          <div key={idx} style={{ display: 'flex', gap: '0.4rem' }}>
            {item.field && <span style={{ fontWeight: 600, color: '#f87171' }}>{item.field}:</span>}
            <span>{item.message || JSON.stringify(item)}</span>
          </div>
        ))}
      </div>
    );
  } else if (displayData) {
    if (typeof displayData === 'object') {
      detailContent = (
        <div style={{ marginTop: '0.4rem', display: 'flex', flexDirection: 'column', gap: '0.2rem', fontSize: '0.8rem', color: '#cbd5e1' }}>
          {Object.entries(displayData).map(([key, value]) => (
            <div key={key} style={{ display: 'flex', gap: '0.4rem' }}>
              <span style={{ fontWeight: 600, color: '#f87171' }}>{key}:</span>
              <span>{typeof value === 'object' ? JSON.stringify(value) : String(value)}</span>
            </div>
          ))}
        </div>
      );
    } else {
      detailContent = (
        <div style={{ marginTop: '0.4rem', fontSize: '0.8rem', color: '#cbd5e1' }}>
          {String(displayData)}
        </div>
      );
    }
  }

  return (
    <div
      style={{
        position: 'fixed',
        top: '20px',
        right: '20px',
        zIndex: 9999,
        maxWidth: '440px',
        width: 'calc(100% - 40px)',
        backgroundColor: '#1e1e24',
        border: '1px solid #ef4444',
        borderLeft: '5px solid #ef4444',
        borderRadius: '8px',
        padding: '1rem 1.25rem',
        color: '#ffffff',
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'space-between',
        gap: '1rem',
        boxShadow: '0 10px 30px rgba(0, 0, 0, 0.5), 0 4px 12px rgba(239, 68, 68, 0.25)',
        boxSizing: 'border-box',
        backdropFilter: 'blur(8px)'
      }}
    >
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.75rem', flex: 1 }}>
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#ef4444"
          strokeWidth="2.2"
          strokeLinecap="round"
          strokeLinejoin="round"
          style={{ flexShrink: 0, marginTop: '2px' }}
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.2rem', flex: 1 }}>
          <span style={{ fontWeight: 700, fontSize: '0.9rem', color: '#ef4444', letterSpacing: '0.02em' }}>
            ERROR OCCURRED
          </span>
          <span style={{ fontWeight: 500, fontSize: '0.9rem', lineHeight: '1.4', color: '#e2e8f0', wordBreak: 'break-word' }}>
            {displayMessage}
          </span>
          {detailContent}
        </div>
      </div>
      {onDismiss && (
        <button
          onClick={onDismiss}
          style={{
            background: 'none',
            border: 'none',
            color: '#94a3b8',
            cursor: 'pointer',
            padding: '4px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: '4px',
            transition: 'color 0.2s'
          }}
          title="Dismiss"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      )}
    </div>
  );
}
