import React from 'react';
import {
  Video, Database, Clock, CloudDownload, Search,
  CheckCircle, AlertTriangle, XCircle, RefreshCw, AlertCircle
} from 'lucide-react';
import { PreflightResponse } from '@/services/preflight';

interface PreflightResultCardProps {
  data: PreflightResponse | null;
  loading: boolean;
  onProceed: () => void;
  proceedLabel?: string;
}

interface MetricCardProps {
  label: string;
  value: string | number;
  icon: React.ReactNode;
  color?: 'emerald' | 'red' | 'amber' | 'slate';
}

function MetricCard({ label, value, icon, color = 'slate' }: MetricCardProps) {
  const colorClasses = {
    emerald: 'text-emerald-700',
    red: 'text-red-700',
    amber: 'text-amber-700',
    slate: 'text-slate-900'
  };

  return (
    <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm flex items-center gap-3">
      <div className={`p-2 rounded-lg bg-slate-50 ${colorClasses[color]}`}>{icon}</div>
      <div>
        <div className="text-xs font-medium text-slate-500 uppercase tracking-wider">{label}</div>
        <div className={`text-lg font-bold ${colorClasses[color]}`}>{value}</div>
      </div>
    </div>
  );
}

export default function PreflightResultCard({
  data,
  loading,
  onProceed,
  proceedLabel = 'Proceed with Operation'
}: PreflightResultCardProps) {
  if (loading) {
    return (
      <div className="p-6 rounded-xl border border-slate-200 bg-white shadow-sm text-slate-600 flex items-center gap-3 my-4">
        <RefreshCw className="animate-spin h-5 w-5 text-blue-600" />
        <span>Running preflight analysis...</span>
      </div>
    );
  }

  if (!data) return null;

  const StatusIcon = data.valid ? CheckCircle : XCircle;

  return (
    <div className={`p-6 rounded-xl border my-4 shadow-md ${data.valid ? 'bg-emerald-50 border-emerald-200' : 'bg-red-50 border-red-200'}`}>
      <div className="flex items-start justify-between mb-6">
        <div className="flex items-center gap-3">
          <StatusIcon className={`h-8 w-8 ${data.valid ? 'text-emerald-600' : 'text-red-600'}`} />
          <div>
            <h3 className={`text-lg font-bold ${data.valid ? 'text-emerald-900' : 'text-red-900'}`}>
              {data.valid ? 'Preflight Check Passed' : 'Preflight Check Failed'}
            </h3>
            {data.error && (
              <p className="text-sm text-red-700 font-medium mt-0.5">{data.error}</p>
            )}
          </div>
        </div>
      </div>

      <div className="overflow-x-auto mb-6">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-slate-300">
              <th className="py-2 px-4 text-sm font-semibold text-slate-700">Metric</th>
              <th className="py-2 px-4 text-sm font-semibold text-slate-700">Value</th>
            </tr>
          </thead>
          <tbody className="text-sm text-slate-800">
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4">Unique Videos</td>
              <td className="py-2 px-4 font-mono">{data.unique_video_count}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4">Total Resources</td>
              <td className="py-2 px-4 font-mono">{data.total_resource_count}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4">Estimated Work</td>
              <td className="py-2 px-4 font-mono">{data.estimated_work}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4">Direct Videos Count</td>
              <td className="py-2 px-4 font-mono">{data.direct_videos?.item_count ?? 0}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4">Total Playlists</td>
              <td className="py-2 px-4 font-mono">{data.playlists?.total_playlists ?? 0}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4">Total Playlist Videos</td>
              <td className="py-2 px-4 font-mono">{data.playlists?.total_videos ?? 0}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4 text-emerald-700">Combined Metadata Cached</td>
              <td className="py-2 px-4 font-mono text-emerald-700">{data.combined?.metadata_cached ?? 0}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4 text-amber-700">Combined Metadata Upstream</td>
              <td className="py-2 px-4 font-mono text-amber-700">{data.combined?.metadata_upstream ?? 0}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4 text-emerald-700">Combined Subtitle Cached</td>
              <td className="py-2 px-4 font-mono text-emerald-700">{data.combined?.subtitle_cached ?? 0}</td>
            </tr>
            <tr className="border-b border-slate-200">
              <td className="py-2 px-4 text-amber-700">Combined Subtitle Upstream</td>
              <td className="py-2 px-4 font-mono text-amber-700">{data.combined?.subtitle_upstream ?? 0}</td>
            </tr>
          </tbody>
        </table>
      </div>

      {data.valid && (
        <div className="flex justify-end">
          <button
            onClick={onProceed}
            className="px-6 py-2.5 rounded-lg text-white font-semibold shadow-sm transition-all bg-emerald-600 hover:bg-emerald-700 hover:shadow-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2"
          >
            {proceedLabel}
          </button>
        </div>
      )}
    </div>
  );
}
