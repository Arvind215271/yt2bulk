import React from 'react';
import { PreflightResponse } from '@/services/preflight';
import { theme } from '@/theme/theme';

interface PreflightResultCardProps {
  data: PreflightResponse | null;
  loading: boolean;
  onProceed: () => void;
  proceedLabel?: string;
}

export default function PreflightResultCard({
  data,
  loading,
  onProceed,
  proceedLabel = 'Proceed with Operation'
}: PreflightResultCardProps) {
  if (loading) {
    return (
      <div className="p-4 rounded-lg border border-gray-200 bg-gray-50 text-gray-600 flex items-center space-x-3 my-4">
        <svg className="animate-spin h-5 w-5 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span>Running preflight analysis...</span>
      </div>
    );
  }

  if (!data) return null;

  return (
    <div className={`p-5 rounded-lg border my-4 shadow-sm ${data.valid ? 'bg-emerald-50 border-emerald-200' : 'bg-red-50 border-red-200'}`}>
      <div className="flex items-start justify-between">
        <div>
          <h3 className={`text-lg font-semibold flex items-center space-x-2 ${data.valid ? 'text-emerald-900' : 'text-red-900'}`}>
            <span>{data.valid ? '✅ Preflight Check Passed' : '❌ Preflight Check Failed'}</span>
          </h3>
          {data.error && (
            <p className="mt-1 text-sm text-red-700 font-medium">{data.error}</p>
          )}
        </div>
      </div>

      <div className="mt-4 grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm">
        <div className="bg-white/80 p-3 rounded border border-gray-100 shadow-xs">
          <div className="text-gray-500 font-medium">Unique Videos</div>
          <div className="text-xl font-bold text-gray-900 mt-1">{data.unique_video_count}</div>
        </div>
        <div className="bg-white/80 p-3 rounded border border-gray-100 shadow-xs">
          <div className="text-gray-500 font-medium">Total Resources</div>
          <div className="text-xl font-bold text-gray-900 mt-1">{data.total_resource_count}</div>
        </div>
        <div className="bg-white/80 p-3 rounded border border-gray-100 shadow-xs">
          <div className="text-gray-500 font-medium">Estimated Work Units</div>
          <div className="text-xl font-bold text-gray-900 mt-1">{data.estimated_work}</div>
        </div>
      </div>

      {data.valid && (
        <div className="mt-5 flex justify-end">
          <button
            onClick={onProceed}
            className="px-5 py-2.5 rounded-md text-white font-medium shadow-sm transition-colors bg-emerald-600 hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2"
          >
            {proceedLabel}
          </button>
        </div>
      )}
    </div>
  );
}
