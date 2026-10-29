'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { theme } from '@/theme/theme';

export default function Header() {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <header style={{
      backgroundColor: theme.colors.panelBg,
      borderBottom: `1px solid ${theme.colors.borderColor}`,
      padding: '1rem 1.5rem',
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      position: 'sticky',
      top: 0,
      zIndex: 50
    }}>
      <Link href="/" style={{
        fontFamily: theme.fonts.mono,
        fontWeight: 800,
        fontSize: '1.25rem',
        color: theme.colors.primaryRed,
        textDecoration: 'none',
        letterSpacing: '-0.05em'
      }}>
        YT2BULK
      </Link>

      {/* Desktop Nav */}
      <nav className="desktop-nav" style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
        <Link href="/" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.9rem', fontWeight: 500 }}>
          Home
        </Link>
        <Link href="/playlist" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.9rem', fontWeight: 500 }}>
          Playlist
        </Link>
        <Link href="/channel" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.9rem', fontWeight: 500 }}>
          Channel
        </Link>
        <Link href="/subtitle" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.9rem', fontWeight: 500 }}>
          Subtitle
        </Link>
        <Link href="/preflight" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.9rem', fontWeight: 500 }}>
          Preflight
        </Link>
        <Link href="/export" style={{ color: theme.colors.primaryRed, textDecoration: 'none', fontSize: '0.9rem', fontWeight: 600 }}>
          Export
        </Link>
        <a href="https://github.com/Arvind215271/yt2bulk" target="_blank" rel="noopener noreferrer" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.9rem', fontWeight: 500 }}>
          GitHub
        </a>
      </nav>

      {/* Mobile Hamburger Button */}
      <button
        className="mobile-menu-btn"
        onClick={() => setMenuOpen(!menuOpen)}
        style={{
          background: 'transparent',
          border: 'none',
          color: theme.colors.textMain,
          fontSize: '1.5rem',
          cursor: 'pointer',
          display: 'none',
          padding: '0.25rem'
        }}
        aria-label="Toggle Navigation Menu"
      >
        {menuOpen ? '✕' : '☰'}
      </button>

      {/* Mobile Menu Dropdown */}
      {menuOpen && (
        <div style={{
          position: 'absolute',
          top: '100%',
          left: 0,
          right: 0,
          backgroundColor: theme.colors.panelBg,
          borderBottom: `1px solid ${theme.colors.borderColor}`,
          padding: '1.25rem 1.5rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '1rem',
          boxShadow: '0 8px 24px rgba(0,0,0,0.4)'
        }} className="mobile-dropdown">
          <Link href="/" onClick={() => setMenuOpen(false)} style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '1rem', fontWeight: 500 }}>
            Home
          </Link>
          <Link href="/playlist" onClick={() => setMenuOpen(false)} style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '1rem', fontWeight: 500 }}>
            Playlist
          </Link>
          <Link href="/channel" onClick={() => setMenuOpen(false)} style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '1rem', fontWeight: 500 }}>
            Channel
          </Link>
          <Link href="/subtitle" onClick={() => setMenuOpen(false)} style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '1rem', fontWeight: 500 }}>
            Subtitle
          </Link>
          <Link href="/preflight" onClick={() => setMenuOpen(false)} style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '1rem', fontWeight: 500 }}>
            Preflight
          </Link>
          <Link href="/export" onClick={() => setMenuOpen(false)} style={{ color: theme.colors.primaryRed, textDecoration: 'none', fontSize: '1rem', fontWeight: 600 }}>
            Export
          </Link>
          <a href="https://github.com/Arvind215271/yt2bulk" target="_blank" rel="noopener noreferrer" onClick={() => setMenuOpen(false)} style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '1rem', fontWeight: 500 }}>
            GitHub
          </a>
        </div>
      )}

      <style jsx>{`
        @media (max-width: 768px) {
          .desktop-nav {
            display: none !important;
          }
          .mobile-menu-btn {
            display: block !important;
          }
        }
        @media (min-width: 769px) {
          .mobile-dropdown {
            display: none !important;
          }
        }
      `}</style>
    </header>
  );
}
