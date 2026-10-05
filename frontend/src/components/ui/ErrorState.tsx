import React from 'react';
import { AlertTriangle, RotateCw } from 'lucide-react';
import { useNetwork } from '../../context/NetworkContext';

interface ErrorStateProps {
  message?: string;
  onRetry?: () => void;
}

export const ErrorState: React.FC<ErrorStateProps> = ({ message, onRetry }) => {
  const { clearError } = useNetwork();

  const handleRetry = () => {
    if (onRetry) {
      onRetry();
    } else {
      clearError();
    }
  };

  return (
    <div className="flex flex-col items-center justify-center p-12 text-center max-w-md mx-auto">
      <div className="w-12 h-12 rounded-full bg-red-100 dark:bg-red-950/60 border border-red-200 dark:border-red-900 flex items-center justify-center text-red-600 dark:text-red-400 mb-4">
        <AlertTriangle className="w-6 h-6" />
      </div>
      <h3 className="text-base font-bold text-neutral-900 dark:text-neutral-100 mb-2">
        Unable to inspect local network
      </h3>
      <p className="text-xs text-neutral-600 dark:text-neutral-400 mb-4 leading-relaxed">
        {message || 'NetWatch could not access the selected network interface or query the local ARP tables.'}
      </p>

      <div className="w-full text-left bg-neutral-100 dark:bg-neutral-800/80 border border-neutral-200 dark:border-neutral-700/80 rounded-md p-3 text-xs text-neutral-600 dark:text-neutral-400 mb-6">
        <div className="font-semibold text-neutral-800 dark:text-neutral-200 mb-1.5">
          Possible causes:
        </div>
        <ul className="space-y-1 list-disc list-inside text-[11px]">
          <li>Network cable unplugged or Wi-Fi radio disconnected</li>
          <li>Elevated administrator permissions required for raw socket access</li>
          <li>Third-party VPN or corporate firewall intercepting ARP broadcasts</li>
          <li>Network adapter driver currently restarting or updating</li>
        </ul>
      </div>

      <button
        onClick={handleRetry}
        className="inline-flex items-center gap-2 px-4 py-2 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-900 rounded text-xs font-semibold hover:bg-neutral-800 dark:hover:bg-white transition-colors"
      >
        <RotateCw className="w-3.5 h-3.5" />
        <span>Try Again</span>
      </button>
    </div>
  );
};
