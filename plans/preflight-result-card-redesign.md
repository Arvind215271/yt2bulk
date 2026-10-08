# PreflightResultCard Redesign Plan

## Objective
Modernize `PreflightResultCard` for better visual hierarchy and scannability using Tailwind CSS and `lucide-react`.

## Layout Specification

### 1. Main Container
- Rounded: `rounded-xl`
- Shadow: `shadow-md`
- Border: `border-slate-200`
- Background: `bg-white`

### 2. Header (Status)
- Icon-based status (e.g., `CheckCircle`, `AlertCircle`, `XCircle`).
- Distinctive background/text color:
    - Success: `bg-emerald-50 text-emerald-800`
    - Failure: `bg-red-50 text-red-800`
- Clear action label.

### 3. Metric Grouping
Divide metrics into two logical sections for better readability.

#### Section A: Scope (Overview)
- **Unique Videos** (`Video` icon)
- **Total Resources** (`Database` icon)
- **Estimated Work** (`Clock` icon)

#### Section B: Technical Health (Detailed Stats)
- **Metadata Upstream Fetches** (`CloudDownload` icon)
- **Total Metadata Fetches** (`Search` icon)
- **Successes/Failures/Fallbacks** (Use colored badges or a small progress-like indicator: `CheckCircle`, `AlertTriangle`, `RefreshCw` icons)

## Visual Hierarchy
- Use bold, larger font sizes for the metric values.
- Use smaller, muted text for labels (`text-slate-500`).
- Use consistent card spacing (`p-6`, `gap-6`).
- Use `lucide-react` icons for visual cues.

## Implementation Details
1. Create a sub-component `MetricCard` to reduce code duplication and ensure consistent styling of metric boxes.
2. Ensure accessibility (semantic HTML).
3. Smooth transitions/loading state.
