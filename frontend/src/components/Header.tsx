import Link from 'next/link';
import { theme } from '@/theme/theme';

export default function Header() {
  return (
    <header style={{
      backgroundColor: theme.colors.panelBg,
      borderBottom: `1px solid ${theme.colors.borderColor}`,
      padding: '1rem',
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      flexWrap: 'wrap',
      gap: '0.5rem',
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
      <nav style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', justifyContent: 'center' }}>
        <Link href="/" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.85rem', fontWeight: 500 }}>
          Home
        </Link>
        <Link href="/about" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.85rem', fontWeight: 500 }}>
          About
        </Link>
        <Link href="/playlist" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.85rem', fontWeight: 500 }}>
          Playlist
        </Link>
        <Link href="/transcript" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.85rem', fontWeight: 500 }}>
          Transcript
        </Link>
        <Link href="/subtitle" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.85rem', fontWeight: 500 }}>
          Subtitle
        </Link>
        <Link href="/preflight" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.85rem', fontWeight: 500 }}>
          Preflight
        </Link>
        <Link href="/export" style={{ color: theme.colors.primaryRed, textDecoration: 'none', fontSize: '0.85rem', fontWeight: 600 }}>
          Export
        </Link>
        <a href="https://github.com/Arvind215271/yt2bulk" target="_blank" rel="noopener noreferrer" style={{ color: theme.colors.textMuted, textDecoration: 'none', fontSize: '0.85rem', fontWeight: 500 }}>
          GitHub
        </a>
      </nav>
    </header>
  );
}
