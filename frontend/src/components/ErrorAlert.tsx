'use client';

import React, { useEffect, useState } from 'react';
import { theme } from '@/theme/theme';

interface ErrorAlertProps {
  message: string | null;
  onDismiss?: () => void;
  autoDismissMs?: number;
}

export default function ErrorAlert({ message, onDismiss, autoDismissMs = 8000 }: ErrorAlertProps) {
  const [fadingOut, setFadingOut] = useState(false);

  useEffect(() => {
    if (!message) {
      setFadingOut(false);
      return;
    }

    if (!autoDismissMs) return;

    // Start fade out slightly before removal
    const fadeTimer = setTimeout(() => {
      setFadingOut(true);
    }, autoDismissMs - 600);

    const dismissTimer = setTimeout(() => {
      if (onDismiss) onDismiss();
      setFadingOut(false);
    }, autoDismissMs);

    return () => {
      clearTimeout(fadeTimer);
      clearTimeout(dismissTimer);
    };
  }, [message, autoDismissMs, onDismiss]);

  if (!message) return null;

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
        border: `1px solid ${theme.colors.primaryRed}`,
        borderLeft: `5px solid ${theme.colors.primaryRed}`,
        borderRadius: '8px',
        padding: '1rem 1.25rem',
        color: '#ffffff',
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'space-between',
        gap: '1rem',
        boxShadow: '0 10px 30px rgba(0, 0, 0, 0.5), 0 4px 12px rgba(239, 68, 68, 0.25)',
        boxSizing: 'border-box',
        opacity: fadingOut ? 0 : 1,
        transform: fadingOut ? 'translateY(-10px)' : 'translateY(0)',
        transition: 'opacity 0.6s ease-in-out, transform 0.6s ease-in-out',
        backdropFilter: 'blur(8px)'
      }}
    >
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.75rem', flex: 1 }}>
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke={theme.colors.primaryRed}
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
          <span style={{ fontWeight: 700, fontSize: '0.9rem', color: theme.colors.primaryRed, letterSpacing: '0.02em' }}>
            ERROR OCCURRED
          </span>
          <span style={{ fontWeight: 500, fontSize: '0.9rem', lineHeight: '1.4', color: '#e2e8f0', wordBreak: 'break-word' }}>
            {message}
          </span>
        </div>
      </div>
      {onDismiss && (
        <button
          onClick={() => {
            setFadingOut(true);
            setTimeout(() => {
              onDismiss();
              setFadingOut(false);
            }, 300);
          }}
          style={{
            background: 'transparent',
            border: 'none',
            color: '#94a3b8',
            cursor: 'pointer',
            fontSize: '1.25rem',
            fontWeight: 700,
            lineHeight: 1,
            padding: '0.2rem',
            borderRadius: '4px',
            transition: 'color 0.2s, background 0.2s'
          }}
          onMouseEnter={(e) => (e.currentTarget.style.color = '#ffffff')}
          onMouseLeave={(e) => (e.currentTarget.style.color = '#94a3b8')}
          title="Dismiss error"
        >
          &times;
        </button>
      )}
    </div>
  );
}
