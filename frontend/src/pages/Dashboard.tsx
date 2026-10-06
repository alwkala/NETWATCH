import React from 'react';
import { useNetwork } from '../context/NetworkContext';
import { MetricCard } from '../components/ui/MetricCard';
import { DeviceTable } from '../components/devices/DeviceTable';
import { DashboardLoadingSkeleton } from '../components/ui/LoadingSkeleton';
import { ErrorState } from '../components/ui/ErrorState';
import {
  RotateCw,
  Radar,
  ArrowRight,
  Activity,
  PlusCircle,
  Radio,
  Clock,
  ExternalLink
} from 'lucide-react';

export const Dashboard: React.FC = () => {
  const {
    devices,
    networkInfo,
    events,
    isScanning,
    scanProgress,
    startScan,
    navigateTo,
    selectDevice,
    isLoading,
    error,
    clearError
  } = useNetwork();

  if (isLoading) {
    return <DashboardLoadingSkeleton />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={clearError} />;
  }

  const onlineCount = devices.filter(d => d.status === 'online').length;
  const offlineCount = devices.filter(d => d.status === 'offline').length;
  const newCount = devices.filter(d => d.isNew).length;

  const avgLatency = networkInfo?.pingStats.avgMs ?? 7;
  const currentLatency = networkInfo?.pingStats.currentMs ?? 4;
  const peakLatency = networkInfo?.pingStats.peakMs ?? 21;
  const latencyHistory = networkInfo?.pingStats.history ?? [4, 5, 7, 4, 6, 8, 21, 6, 5, 4, 4, 7, 5, 4, 6];

  // Empty state check
  if (devices.length === 0) {
    return (
      <div className="p-8 max-w-4xl mx-auto flex flex-col items-center justify-center min-h-[60vh] text-center">
        <div className="w-16 h-16 rounded-full bg-neutral-100 dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-700 flex items-center justify-center text-neutral-400 mb-4">
          <Radio className="w-8 h-8" />
        </div>
        <h2 className="text-xl font-bold text-neutral-900 dark:text-neutral-100 mb-2">
          No devices discovered
        </h2>
        <p className="text-xs text-neutral-500 dark:text-neutral-400 max-w-md mb-6 leading-relaxed">
          Run a network scan to discover all active devices connected to your local network ({networkInfo?.subnet || '192.168.1.0/24'}).
        </p>
        <button
          onClick={() => startScan('quick')}
          disabled={isScanning}
          className="px-5 py-2.5 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-950 text-xs font-semibold rounded hover:bg-neutral-800 dark:hover:bg-white transition-colors flex items-center gap-2 shadow-xs"
        >
          {isScanning ? (
            <>
              <RotateCw className="w-4 h-4 animate-spin" />
              <span>Scanning... {scanProgress.progress}%</span>
            </>
          ) : (
            <>
              <Radar className="w-4 h-4 text-emerald-500" />
              <span>Scan Network</span>
            </>
          )}
        </button>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header section (Section 8) */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-1 border-b border-neutral-200/80 dark:border-neutral-800">
        <div>
          <div className="flex items-center gap-2.5 mb-1">
            <h2 className="text-xl font-bold tracking-tight text-neutral-900 dark:text-neutral-100">
              {networkInfo?.networkName || 'Home Network'}
            </h2>
            <span className="font-mono text-xs px-2 py-0.5 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-600 dark:text-neutral-300 border border-neutral-200 dark:border-neutral-700">
              {networkInfo?.subnet || '192.168.1.0/24'}
            </span>
          </div>
          <div className="flex items-center gap-2 text-xs text-neutral-500 dark:text-neutral-400">
            <span className="inline-flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400 font-medium">
              <span className="w-2 h-2 rounded-full bg-emerald-500" />
              Connected
            </span>
            <span>·</span>
            <span>Gateway: <strong className="font-mono font-medium text-neutral-700 dark:text-neutral-300">{networkInfo?.gateway || '192.168.1.1'}</strong></span>
            <span>·</span>
            <span>Host: <strong className="font-mono font-medium text-neutral-700 dark:text-neutral-300">{networkInfo?.localIp || '192.168.1.24'}</strong></span>
          </div>
        </div>
      </div>

      {/* Live Scanning Progress Banner (if active) */}
      {isScanning && (
        <div className="p-4 rounded-md bg-neutral-100 dark:bg-neutral-800/90 border border-neutral-300 dark:border-neutral-700 space-y-2">
          <div className="flex justify-between items-center text-xs font-medium">
            <span className="flex items-center gap-2 text-neutral-800 dark:text-neutral-200">
              <RotateCw className="w-3.5 h-3.5 animate-spin text-emerald-500" />
              Scanning subnet {networkInfo?.subnet || '192.168.1.0/24'}...
            </span>
            <span className="font-mono tabular-nums text-neutral-600 dark:text-neutral-400">
              {scanProgress.scanned} / {scanProgress.total} addresses ({scanProgress.progress}%)
            </span>
          </div>
          {/* Progress bar */}
          <div className="w-full bg-neutral-200 dark:bg-neutral-700 h-2 rounded-full overflow-hidden">
            <div
              className="bg-emerald-500 h-full transition-all duration-200"
              style={{ width: `${scanProgress.progress}%` }}
            />
          </div>
          <div className="text-[11px] text-neutral-500 dark:text-neutral-400 flex justify-between">
            <span>Devices identified: <strong className="font-mono">{scanProgress.found}</strong></span>
            <span>ARP + ICMP sweep in progress</span>
          </div>
        </div>
      )}

      {/* KPI Cards (Section 8) */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
        <MetricCard
          label="Devices"
          value={devices.length}
          subtext="discovered"
          onClick={() => navigateTo('devices')}
        />
        <MetricCard
          label="Online"
          value={onlineCount}
          subtext="active now"
          highlight="online"
          onClick={() => navigateTo('devices')}
        />
        <MetricCard
          label="Offline"
          value={offlineCount}
          subtext="dormant"
          highlight="offline"
          onClick={() => navigateTo('devices')}
        />
        <MetricCard
          label="New"
          value={newCount}
          subtext={newCount > 0 ? 'review needed' : 'none'}
          highlight="new"
          onClick={() => navigateTo('devices')}
        />
        <MetricCard
          label="Avg Latency"
          value={`${avgLatency} ms`}
          subtext={`peak ${peakLatency} ms`}
          highlight="accent"
        />
      </div>

      {/* Main Grid: Device Overview (Left 2 cols) & Side widgets (Right 1 col) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Left Column: Device Overview Table */}
        <div className="lg:col-span-2 space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
                Devices
              </h3>
              <p className="text-xs text-neutral-500 dark:text-neutral-400">
                {devices.length} devices discovered on {networkInfo?.subnet || '192.168.1.0/24'}
              </p>
            </div>
            <button
              onClick={() => navigateTo('devices')}
              className="text-xs font-semibold text-neutral-700 dark:text-neutral-300 hover:text-neutral-900 dark:hover:text-white flex items-center gap-1 group"
            >
              <span>View all {devices.length} devices</span>
              <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform" />
            </button>
          </div>

          {/* Compact Device Table (shows top 8 rows on dashboard) */}
          <DeviceTable
            devices={devices}
            limit={8}
            compact={true}
            onSelectDevice={dev => selectDevice(dev.id)}
            showControls={false}
          />
        </div>

        {/* Right Column: Recent Activity & Network Health */}
        <div className="space-y-6">
          {/* Network Health Widget (Section 11) */}
          <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-4 space-y-3 shadow-xs transition-colors">
            <div className="flex items-center justify-between">
              <span className="text-[11px] font-semibold tracking-wider text-neutral-500 dark:text-neutral-400 uppercase">
                Network Health
              </span>
              <span className="text-[11px] text-emerald-600 dark:text-emerald-400 font-medium">
                Optimal
              </span>
            </div>

            {/* Latency sparkline / distribution */}
            <div>
              <div className="flex items-center justify-between text-xs text-neutral-500 mb-1.5">
                <span>ICMP Latency distribution</span>
                <span className="font-mono text-neutral-800 dark:text-neutral-200">{currentLatency} ms</span>
              </div>
              <div className="flex items-end gap-1 h-10 py-1 bg-neutral-50 dark:bg-neutral-800 rounded px-2">
                {latencyHistory.map((val, idx) => {
                  const heightPercent = Math.min(100, Math.max(15, (val / peakLatency) * 100));
                  return (
                    <div
                      key={idx}
                      className="flex-1 bg-neutral-300 dark:bg-neutral-600 hover:bg-emerald-500 transition-colors rounded-xs"
                      style={{ height: `${heightPercent}%` }}
                      title={`${val} ms`}
                    />
                  );
                })}
              </div>
            </div>

            {/* Metrics breakdown */}
            <div className="grid grid-cols-3 pt-2 border-t border-neutral-100 dark:border-neutral-800 text-xs">
              <div>
                <div className="text-[10px] text-neutral-400 uppercase">Current</div>
                <div className="font-mono font-semibold text-neutral-800 dark:text-neutral-200 tabular-nums">
                  {currentLatency} ms
                </div>
              </div>
              <div>
                <div className="text-[10px] text-neutral-400 uppercase">Average</div>
                <div className="font-mono font-semibold text-neutral-800 dark:text-neutral-200 tabular-nums">
                  {avgLatency} ms
                </div>
              </div>
              <div>
                <div className="text-[10px] text-neutral-400 uppercase">Peak</div>
                <div className="font-mono font-semibold text-neutral-800 dark:text-neutral-200 tabular-nums">
                  {peakLatency} ms
                </div>
              </div>
            </div>
          </div>

          {/* Recent Events (Section 10) */}
          <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-4 space-y-3 shadow-xs transition-colors">
            <div className="flex items-center justify-between">
              <span className="text-[11px] font-semibold tracking-wider text-neutral-500 dark:text-neutral-400 uppercase">
                Recent Activity
              </span>
              <button
                onClick={() => navigateTo('events')}
                className="text-xs text-neutral-500 dark:text-neutral-400 hover:text-neutral-800 dark:hover:text-neutral-200"
              >
                View all →
              </button>
            </div>

            <div className="space-y-3">
              {events.slice(0, 4).map(evt => (
                <div
                  key={evt.id}
                  onClick={() => {
                    if (evt.deviceId) {
                      selectDevice(evt.deviceId);
                    } else {
                      navigateTo('events');
                    }
                  }}
                  className="flex items-start gap-2.5 text-xs cursor-pointer hover:bg-neutral-50 dark:hover:bg-neutral-800 p-1.5 rounded transition-colors group"
                >
                  <span
                    className={`w-2 h-2 rounded-full mt-1 shrink-0 ${
                      evt.type === 'new_device'
                        ? 'bg-amber-500'
                        : evt.type === 'online'
                        ? 'bg-emerald-500'
                        : evt.type === 'offline'
                        ? 'bg-neutral-400 dark:bg-neutral-500'
                        : 'bg-sky-500'
                    }`}
                  />
                  <div className="flex-1 min-w-0">
                    <div className="font-semibold text-neutral-900 dark:text-neutral-100 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                      {evt.title}
                    </div>
                    {evt.deviceName && (
                      <div className="text-neutral-600 dark:text-neutral-400 truncate">
                        {evt.deviceName} {evt.ip && <span className="font-mono text-[11px] text-neutral-400">({evt.ip})</span>}
                      </div>
                    )}
                    <div className="text-[10px] text-neutral-400 mt-0.5">
                      {evt.timestamp}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
