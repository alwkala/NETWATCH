import React, { useState } from 'react';
import { useNetwork, PrototypeStatePreset } from '../../context/NetworkContext';
import {
  Activity,
  Minus,
  Square,
  X,
  SlidersHorizontal,
  ChevronDown
} from 'lucide-react';

export const TitleBar: React.FC = () => {
  const { prototypeState, setPrototypeStatePreset, service } = useNetwork();
  const simulated = service.isSimulated;
  const [showPresetMenu, setShowPresetMenu] = useState(false);
  const [isMaximized, setIsMaximized] = useState(false);

  const presets: { id: PrototypeStatePreset; label: string; desc: string }[] = [
    { id: 'normal', label: 'Normal Dashboard', desc: 'Default populated state (24 devices)' },
    { id: 'scanning', label: 'Scanning Active', desc: 'Simulated live discovery sweep' },
    { id: 'scan_completed', label: 'Scan Completed', desc: 'Results report & device tally' },
    { id: 'new_device', label: 'New Device Detected', desc: 'Inspect unverified Xiaomi device' },
    { id: 'device_offline', label: 'Device Offline', desc: 'Inspect disconnected ThinkPad' },
    { id: 'empty', label: 'Empty Network', desc: 'Zero devices discovered state' },
    { id: 'loading', label: 'Loading Skeleton', desc: 'Data fetching skeleton preview' },
    { id: 'error', label: 'Interface Error', desc: 'Simulated interface access failure' }
  ];

  return (
    <header className="h-9 select-none bg-neutral-100 dark:bg-neutral-900 border-b border-neutral-200 dark:border-neutral-800 flex items-center justify-between px-3 text-xs z-50 text-neutral-600 dark:text-neutral-400">
      {/* Left: App Identity */}
      <div className="flex items-center gap-2">
        <div className="w-4 h-4 rounded-xs bg-neutral-900 dark:bg-neutral-100 flex items-center justify-center text-white dark:text-neutral-950 font-bold">
          <Activity className="w-3 h-3 text-emerald-500 dark:text-emerald-600" />
        </div>
        <span className="font-semibold text-neutral-800 dark:text-neutral-200 tracking-tight">
          NETWATCH
        </span>
        <span className="text-neutral-400 dark:text-neutral-600">·</span>
        <span className="text-neutral-500 dark:text-neutral-400 hidden sm:inline">
          Local Network Intelligence
        </span>
      </div>

      {/* Center: Prototype State Tester Switcher (prototype only) */}
      {!simulated ? <div /> : (
      <div className="relative">
        <button
          onClick={() => setShowPresetMenu(!showPresetMenu)}
          className="flex items-center gap-1.5 px-2 py-1 rounded text-[11px] font-medium bg-neutral-200/80 dark:bg-neutral-800 hover:bg-neutral-300 dark:hover:bg-neutral-700 text-neutral-700 dark:text-neutral-300 border border-neutral-300/60 dark:border-neutral-700 transition-colors"
          title="Switch prototype scenario states"
        >
          <SlidersHorizontal className="w-3 h-3 text-neutral-500" />
          <span className="text-neutral-500 dark:text-neutral-400">State:</span>
          <span className="font-semibold text-neutral-900 dark:text-neutral-100 capitalize">
            {prototypeState.replace('_', ' ')}
          </span>
          <ChevronDown className="w-3 h-3 opacity-60" />
        </button>

        {showPresetMenu && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setShowPresetMenu(false)}
            />
            <div className="absolute left-1/2 -translate-x-1/2 mt-1 w-64 bg-white dark:bg-neutral-900 rounded-md shadow-xl border border-neutral-200 dark:border-neutral-700 py-1.5 z-50 text-left">
              <div className="px-3 py-1 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider border-b border-neutral-100 dark:border-neutral-800">
                Prototype State Presets (Sec. 34)
              </div>
              {presets.map(p => (
                <button
                  key={p.id}
                  onClick={() => {
                    setPrototypeStatePreset(p.id);
                    setShowPresetMenu(false);
                  }}
                  className={`w-full px-3 py-1.5 text-left text-xs flex flex-col hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors ${
                    prototypeState === p.id
                      ? 'bg-neutral-100/80 dark:bg-neutral-800/80 font-medium text-emerald-600 dark:text-emerald-400'
                      : 'text-neutral-700 dark:text-neutral-300'
                  }`}
                >
                  <span className="text-xs">{p.label}</span>
                  <span className="text-[10px] text-neutral-400 dark:text-neutral-500 font-normal">
                    {p.desc}
                  </span>
                </button>
              ))}
            </div>
          </>
        )}
      </div>
      )}

      {/* Right: Windows Window Controls (simulated prototype only) */}
      <div className="flex items-center gap-1 -mr-2">

        {/* Window controls are only drawn in the browser prototype; the Wails host has a native frame */}
        {simulated && (
        <div className="flex items-center">
          <button
            className="w-10 h-7 flex items-center justify-center hover:bg-neutral-200 dark:hover:bg-neutral-800 text-neutral-500 transition-colors"
            title="Minimize"
            aria-label="Minimize window"
          >
            <Minus className="w-3.5 h-3.5" />
          </button>
          <button
            onClick={() => setIsMaximized(!isMaximized)}
            className="w-10 h-7 flex items-center justify-center hover:bg-neutral-200 dark:hover:bg-neutral-800 text-neutral-500 transition-colors"
            title={isMaximized ? 'Restore' : 'Maximize'}
            aria-label="Maximize window"
          >
            <Square className="w-3 h-3" />
          </button>
          <button
            className="w-10 h-7 flex items-center justify-center hover:bg-red-600 hover:text-white text-neutral-500 transition-colors"
            title="Close"
            aria-label="Close window"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
        )}
      </div>
    </header>
  );
};
