import React, { useState } from 'react';
import { useNetwork } from '../context/NetworkContext';
import {
  Radar,
  RotateCw,
  Search,
  CheckCircle2,
  HardDrive,
  Shield,
  ArrowRight,
  Radio,
  Sliders
} from 'lucide-react';

export const Scanner: React.FC = () => {
  const {
    networkInfo,
    devices,
    isScanning,
    scanProgress,
    lastScanResult,
    startScan,
    navigateTo,
    selectDevice
  } = useNetwork();

  const [scanType, setScanType] = useState<'quick' | 'full'>('quick');

  const handleStartScan = () => {
    startScan(scanType);
  };

  return (
    <div className="p-6 space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-neutral-200 dark:border-neutral-800">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-bold text-neutral-900 dark:text-neutral-100">
              Network Scanner
            </h2>
            <span className="font-mono text-xs px-2 py-0.5 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-600 dark:text-neutral-300 border border-neutral-200 dark:border-neutral-700">
              {networkInfo?.subnet || '192.168.1.0/24'}
            </span>
          </div>
          <p className="text-xs text-neutral-500 dark:text-neutral-400">
            Probe active subnet for connected devices, ARP registrations, and hostnames.
          </p>
        </div>

        <div>
          <button
            onClick={handleStartScan}
            disabled={isScanning}
            className="px-5 py-2 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-950 text-xs font-semibold rounded hover:bg-neutral-800 dark:hover:bg-white transition-colors flex items-center gap-2 shadow-xs disabled:opacity-50"
          >
            {isScanning ? (
              <>
                <RotateCw className="w-3.5 h-3.5 animate-spin" />
                <span>Scanning Subnet...</span>
              </>
            ) : (
              <>
                <Radar className="w-3.5 h-3.5 text-emerald-500" />
                <span>Scan Network</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* Scan Options (Section 20) */}
      <div className="bg-white dark:bg-neutral-850 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-3 shadow-xs">
        <div className="text-xs font-bold text-neutral-900 dark:text-neutral-100">
          Discovery Mode
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <label
            onClick={() => setScanType('quick')}
            className={`p-3.5 rounded-md border text-xs cursor-pointer flex items-start gap-3 transition-colors ${
              scanType === 'quick'
                ? 'bg-neutral-100/90 dark:bg-neutral-800 border-neutral-400 dark:border-neutral-600 ring-1 ring-neutral-400 dark:ring-neutral-600'
                : 'bg-white dark:bg-neutral-850 border-neutral-200 dark:border-neutral-800 hover:bg-neutral-50 dark:hover:bg-neutral-800/40'
            }`}
          >
            <input
              type="radio"
              name="scanType"
              checked={scanType === 'quick'}
              onChange={() => setScanType('quick')}
              className="mt-0.5"
            />
            <div>
              <div className="font-bold text-neutral-900 dark:text-neutral-100">
                Quick Discovery (Recommended)
              </div>
              <div className="text-neutral-500 text-[11px] mt-0.5 leading-relaxed">
                ARP cache query and ICMP echo sweep across /24 subnet. Fast response (~1.5s), minimal network traffic.
              </div>
            </div>
          </label>

          <label
            onClick={() => setScanType('full')}
            className={`p-3.5 rounded-md border text-xs cursor-pointer flex items-start gap-3 transition-colors ${
              scanType === 'full'
                ? 'bg-neutral-100/90 dark:bg-neutral-800 border-neutral-400 dark:border-neutral-600 ring-1 ring-neutral-400 dark:ring-neutral-600'
                : 'bg-white dark:bg-neutral-850 border-neutral-200 dark:border-neutral-800 hover:bg-neutral-50 dark:hover:bg-neutral-800/40'
            }`}
          >
            <input
              type="radio"
              name="scanType"
              checked={scanType === 'full'}
              onChange={() => setScanType('full')}
              className="mt-0.5"
            />
            <div>
              <div className="font-bold text-neutral-900 dark:text-neutral-100">
                Full Discovery & Inspection
              </div>
              <div className="text-neutral-500 text-[11px] mt-0.5 leading-relaxed">
                Deep fingerprinting, DNS reverse-lookups, mDNS Bonjour broadcast, and standard service checks.
              </div>
            </div>
          </label>
        </div>
      </div>

      {/* Scanner State Display */}
      {isScanning ? (
        /* During Scan State (Section 19) */
        <div className="bg-white dark:bg-neutral-850 border border-neutral-300 dark:border-neutral-700 rounded-lg p-6 space-y-4 shadow-xs">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <RotateCw className="w-5 h-5 text-emerald-500 animate-spin" />
              <div>
                <h3 className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
                  Scanning network...
                </h3>
                <p className="text-xs font-mono text-neutral-500">
                  Target: {networkInfo?.subnet || '192.168.1.0/24'}
                </p>
              </div>
            </div>
            <span className="font-mono text-base font-bold text-neutral-900 dark:text-neutral-100 tabular-nums">
              {scanProgress.progress}%
            </span>
          </div>

          {/* Progress bar */}
          <div className="w-full bg-neutral-200 dark:bg-neutral-700 h-2.5 rounded-full overflow-hidden">
            <div
              className="bg-emerald-500 h-full transition-all duration-150"
              style={{ width: `${scanProgress.progress}%` }}
            />
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 pt-2 text-xs">
            <div className="p-3 bg-neutral-50 dark:bg-neutral-800 rounded">
              <span className="text-[10px] text-neutral-400 uppercase">Addresses Scanned</span>
              <div className="font-mono text-base font-semibold text-neutral-900 dark:text-neutral-100">
                {scanProgress.scanned} / {scanProgress.total}
              </div>
            </div>
            <div className="p-3 bg-neutral-50 dark:bg-neutral-800 rounded">
              <span className="text-[10px] text-neutral-400 uppercase">Devices Discovered</span>
              <div className="font-mono text-base font-semibold text-emerald-600 dark:text-emerald-400">
                {scanProgress.found}
              </div>
            </div>
            <div className="p-3 bg-neutral-50 dark:bg-neutral-800 rounded">
              <span className="text-[10px] text-neutral-400 uppercase">Errors</span>
              <div className="font-mono text-base font-semibold text-neutral-600 dark:text-neutral-400">
                0
              </div>
            </div>
          </div>
        </div>
      ) : lastScanResult ? (
        /* Scan Completed State (Section 19) */
        <div className="bg-white dark:bg-neutral-850 border border-neutral-200 dark:border-neutral-800 rounded-lg p-6 space-y-5 shadow-xs">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-emerald-100 dark:bg-emerald-950 flex items-center justify-center text-emerald-600 dark:text-emerald-400">
              <CheckCircle2 className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
                Scan complete
              </h3>
              <p className="text-xs text-neutral-500">
                Completed at {lastScanResult.timestamp} ({lastScanResult.durationMs}ms)
              </p>
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
            <div className="p-3 bg-neutral-50 dark:bg-neutral-800 rounded">
              <div className="text-[10px] text-neutral-400 uppercase">Scanned</div>
              <div className="text-base font-bold font-mono text-neutral-900 dark:text-neutral-100">
                {lastScanResult.scannedAddresses}
              </div>
              <div className="text-[10px] text-neutral-400">addresses queried</div>
            </div>

            <div className="p-3 bg-neutral-50 dark:bg-neutral-800 rounded">
              <div className="text-[10px] text-neutral-400 uppercase">Found</div>
              <div className="text-base font-bold font-mono text-emerald-600 dark:text-emerald-400">
                {lastScanResult.devicesFound}
              </div>
              <div className="text-[10px] text-neutral-400">devices responding</div>
            </div>

            <div className="p-3 bg-neutral-50 dark:bg-neutral-800 rounded">
              <div className="text-[10px] text-neutral-400 uppercase">New</div>
              <div className="text-base font-bold font-mono text-amber-600 dark:text-amber-400">
                {lastScanResult.newDevices}
              </div>
              <div className="text-[10px] text-neutral-400">new device discovered</div>
            </div>

            <div className="p-3 bg-neutral-50 dark:bg-neutral-800 rounded">
              <div className="text-[10px] text-neutral-400 uppercase">Errors</div>
              <div className="text-base font-bold font-mono text-neutral-600 dark:text-neutral-400">
                0
              </div>
              <div className="text-[10px] text-neutral-400">0 packets dropped</div>
            </div>
          </div>

          <div className="flex gap-3">
            <button
              onClick={() => navigateTo('devices')}
              className="px-4 py-2 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-950 text-xs font-semibold rounded hover:bg-neutral-800 dark:hover:bg-white transition-colors flex items-center gap-1.5"
            >
              <span>View Devices</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={handleStartScan}
              className="px-4 py-2 bg-neutral-100 dark:bg-neutral-800 text-neutral-800 dark:text-neutral-200 text-xs font-medium rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors border border-neutral-300 dark:border-neutral-700"
            >
              Scan Again
            </button>
          </div>
        </div>
      ) : (
        /* Ready to Scan State (Section 19) */
        <div className="bg-white dark:bg-neutral-850 border border-neutral-200 dark:border-neutral-800 rounded-lg p-8 text-center space-y-3 shadow-xs">
          <div className="w-12 h-12 rounded-full bg-neutral-100 dark:bg-neutral-800 flex items-center justify-center text-neutral-500 mx-auto">
            <Radio className="w-6 h-6" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
              Ready to scan
            </h3>
            <p className="text-xs text-neutral-500 mt-1 max-w-sm mx-auto">
              Subnet <span className="font-mono text-neutral-700 dark:text-neutral-300">{networkInfo?.subnet || '192.168.1.0/24'}</span> contains 254 host addresses available for discovery.
            </p>
          </div>
          <button
            onClick={handleStartScan}
            className="px-5 py-2.5 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-950 text-xs font-semibold rounded hover:bg-neutral-800 dark:hover:bg-white transition-colors inline-flex items-center gap-2"
          >
            <Radar className="w-4 h-4 text-emerald-500" />
            <span>Start Quick Discovery</span>
          </button>
        </div>
      )}
    </div>
  );
};
