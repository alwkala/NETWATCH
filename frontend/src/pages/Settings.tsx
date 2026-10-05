import React, { useState } from 'react';
import { useNetwork } from '../context/NetworkContext';
import { useTheme } from '../context/ThemeContext';
import {
  ShieldCheck,
  HardDrive,
  Bell,
  Cpu,
  RefreshCw,
  Check,
  Trash2,
  FolderOpen
} from 'lucide-react';

export const Settings: React.FC = () => {
  const { setPrototypeStatePreset, service, refresh } = useNetwork();
  const { theme, setTheme } = useTheme();

  // General settings state
  const [launchAtStartup, setLaunchAtStartup] = useState(false);
  const [minimizeToTray, setMinimizeToTray] = useState(true);
  const [startMinimized, setStartMinimized] = useState(false);

  // Scanning settings state
  const [autoDiscovery, setAutoDiscovery] = useState(true);
  const [scanInterval, setScanInterval] = useState('5m');

  // Notifications state
  const [notifyNewDevice, setNotifyNewDevice] = useState(true);
  const [notifyDeviceOffline, setNotifyDeviceOffline] = useState(false);
  const [notifyNetworkChange, setNotifyNetworkChange] = useState(true);

  // Feedback states
  const [resetMessage, setResetMessage] = useState<string | null>(null);

  const handleResetData = async () => {
    try {
      if (service.isSimulated || !service.clearHistory) {
        await setPrototypeStatePreset('normal');
      } else {
        await service.clearHistory();
        await refresh();
      }
      setResetMessage('Local device records and event history were cleared.');
    } catch (e) {
      setResetMessage(e instanceof Error ? e.message : 'Could not clear history.');
    }
    setTimeout(() => setResetMessage(null), 3500);
  };

  const handleOpenData = async () => {
    try {
      await service.openDataFolder?.();
    } catch (e) {
      setResetMessage(e instanceof Error ? e.message : 'Could not open the data folder.');
      setTimeout(() => setResetMessage(null), 3500);
    }
  };

  return (
    <div className="p-6 space-y-6 max-w-3xl mx-auto">
      {/* Header */}
      <div className="pb-3 border-b border-neutral-200 dark:border-neutral-800">
        <h2 className="text-lg font-bold text-neutral-900 dark:text-neutral-100">
          Settings & Local Privacy
        </h2>
        <p className="text-xs text-neutral-500 dark:text-neutral-400">
          Configure background discovery, notifications, and review local data storage.
        </p>
      </div>

      {/* PRIVACY SECTION (Mandatory trust-building element - Section 22) */}
      <div className="bg-white dark:bg-neutral-850 border-2 border-emerald-500/30 dark:border-emerald-500/40 rounded-lg p-5 space-y-4 shadow-xs">
        <div className="flex items-center gap-2 text-sm font-bold text-neutral-900 dark:text-neutral-100">
          <ShieldCheck className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
          <span>Local-First & Privacy Guarantee</span>
        </div>

        <p className="text-xs text-neutral-600 dark:text-neutral-300 leading-relaxed font-medium">
          Your network data stays strictly on this computer. NetWatch operates 100% locally with zero cloud dependencies.
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
          <div className="flex items-center gap-2 text-emerald-700 dark:text-emerald-400 font-medium">
            <Check className="w-4 h-4 shrink-0" />
            <span>No account required</span>
          </div>
          <div className="flex items-center gap-2 text-emerald-700 dark:text-emerald-400 font-medium">
            <Check className="w-4 h-4 shrink-0" />
            <span>No cloud storage or syncing</span>
          </div>
          <div className="flex items-center gap-2 text-emerald-700 dark:text-emerald-400 font-medium">
            <Check className="w-4 h-4 shrink-0" />
            <span>No analytics or crash reports</span>
          </div>
          <div className="flex items-center gap-2 text-emerald-700 dark:text-emerald-400 font-medium">
            <Check className="w-4 h-4 shrink-0" />
            <span>No background telemetry</span>
          </div>
          <div className="flex items-center gap-2 text-emerald-700 dark:text-emerald-400 font-medium sm:col-span-2">
            <Check className="w-4 h-4 shrink-0" />
            <span>No external vendor or network database lookup</span>
          </div>
        </div>

        <div className="pt-2 border-t border-neutral-100 dark:border-neutral-800 text-xs">
          <div className="text-[11px] text-neutral-400 mb-1 font-semibold uppercase tracking-wider">
            Local SQLite Source of Truth
          </div>
          <div className="p-2.5 rounded bg-neutral-100 dark:bg-neutral-900 font-mono text-[11px] text-neutral-700 dark:text-neutral-300 flex items-center justify-between break-all">
            <span>C:\Users\User\AppData\Local\NetWatch\data\network.db</span>
            <span className="text-[10px] text-neutral-400 shrink-0 ml-2">WAL Mode</span>
          </div>
        </div>
      </div>

      {/* GENERAL (Section 22) */}
      <div className="bg-white dark:bg-neutral-850 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs">
        <div className="text-xs font-bold text-neutral-900 dark:text-neutral-100 uppercase tracking-wider">
          General
        </div>

        <div className="space-y-3 text-xs">
          <div className="flex items-center justify-between py-1">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                Launch at startup
              </div>
              <div className="text-neutral-400 text-[11px]">
                Start NetWatch minimized in system tray on Windows boot
              </div>
            </div>
            <button
              onClick={() => setLaunchAtStartup(!launchAtStartup)}
              className={`w-11 h-6 rounded-full transition-colors relative p-0.5 ${
                launchAtStartup ? 'bg-neutral-900 dark:bg-neutral-100' : 'bg-neutral-300 dark:bg-neutral-700'
              }`}
            >
              <div
                className={`w-5 h-5 rounded-full bg-white dark:bg-neutral-900 transition-transform ${
                  launchAtStartup ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="flex items-center justify-between py-1 border-t border-neutral-100 dark:border-neutral-800">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                Minimize to tray
              </div>
              <div className="text-neutral-400 text-[11px]">
                Closing the window will minimize to system notification area
              </div>
            </div>
            <button
              onClick={() => setMinimizeToTray(!minimizeToTray)}
              className={`w-11 h-6 rounded-full transition-colors relative p-0.5 ${
                minimizeToTray ? 'bg-neutral-900 dark:bg-neutral-100' : 'bg-neutral-300 dark:bg-neutral-700'
              }`}
            >
              <div
                className={`w-5 h-5 rounded-full bg-white dark:bg-neutral-900 transition-transform ${
                  minimizeToTray ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="flex items-center justify-between py-1 border-t border-neutral-100 dark:border-neutral-800">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                Start minimized
              </div>
              <div className="text-neutral-400 text-[11px]">
                Do not show main window on initial launch
              </div>
            </div>
            <button
              onClick={() => setStartMinimized(!startMinimized)}
              className={`w-11 h-6 rounded-full transition-colors relative p-0.5 ${
                startMinimized ? 'bg-neutral-900 dark:bg-neutral-100' : 'bg-neutral-300 dark:bg-neutral-700'
              }`}
            >
              <div
                className={`w-5 h-5 rounded-full bg-white dark:bg-neutral-900 transition-transform ${
                  startMinimized ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="flex items-center justify-between py-1 border-t border-neutral-100 dark:border-neutral-800">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                Theme Appearance
              </div>
              <div className="text-neutral-400 text-[11px]">
                Select interface display mode
              </div>
            </div>
            <div className="flex gap-1 bg-neutral-100 dark:bg-neutral-800 p-0.5 rounded border border-neutral-200 dark:border-neutral-700">
              <button
                onClick={() => setTheme('light')}
                className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
                  theme === 'light'
                    ? 'bg-white text-neutral-900 shadow-xs'
                    : 'text-neutral-500 hover:text-neutral-800'
                }`}
              >
                Light
              </button>
              <button
                onClick={() => setTheme('dark')}
                className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
                  theme === 'dark'
                    ? 'bg-neutral-700 text-white shadow-xs'
                    : 'text-neutral-400 hover:text-neutral-200'
                }`}
              >
                Dark
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* SCANNING SETTINGS (Section 22) */}
      <div className="bg-white dark:bg-neutral-850 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs">
        <div className="text-xs font-bold text-neutral-900 dark:text-neutral-100 uppercase tracking-wider">
          Scanning & Discovery
        </div>

        <div className="space-y-3 text-xs">
          <div className="flex items-center justify-between py-1">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                Automatic discovery
              </div>
              <div className="text-neutral-400 text-[11px]">
                Continuously listen to ARP broadcasts and passive NDP packets
              </div>
            </div>
            <button
              onClick={() => setAutoDiscovery(!autoDiscovery)}
              className={`w-11 h-6 rounded-full transition-colors relative p-0.5 ${
                autoDiscovery ? 'bg-neutral-900 dark:bg-neutral-100' : 'bg-neutral-300 dark:bg-neutral-700'
              }`}
            >
              <div
                className={`w-5 h-5 rounded-full bg-white dark:bg-neutral-900 transition-transform ${
                  autoDiscovery ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="flex items-center justify-between py-1 border-t border-neutral-100 dark:border-neutral-800">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                Scan interval
              </div>
              <div className="text-neutral-400 text-[11px]">
                Periodic active ICMP sweep frequency
              </div>
            </div>
            <select
              value={scanInterval}
              onChange={e => setScanInterval(e.target.value)}
              className="bg-white dark:bg-neutral-800 text-xs border border-neutral-300 dark:border-neutral-700 rounded px-2.5 py-1 text-neutral-800 dark:text-neutral-200"
            >
              <option value="1m">1 minute</option>
              <option value="5m">5 minutes (Default)</option>
              <option value="15m">15 minutes</option>
              <option value="30m">30 minutes</option>
              <option value="1h">1 hour</option>
            </select>
          </div>
        </div>
      </div>

      {/* NOTIFICATIONS (Section 22) */}
      <div className="bg-white dark:bg-neutral-850 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs">
        <div className="text-xs font-bold text-neutral-900 dark:text-neutral-100 uppercase tracking-wider">
          Windows Notifications
        </div>

        <div className="space-y-3 text-xs">
          <div className="flex items-center justify-between py-1">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                New device detected
              </div>
              <div className="text-neutral-400 text-[11px]">
                Notify when an unrecognized MAC address connects to subnet
              </div>
            </div>
            <button
              onClick={() => setNotifyNewDevice(!notifyNewDevice)}
              className={`w-11 h-6 rounded-full transition-colors relative p-0.5 ${
                notifyNewDevice ? 'bg-neutral-900 dark:bg-neutral-100' : 'bg-neutral-300 dark:bg-neutral-700'
              }`}
            >
              <div
                className={`w-5 h-5 rounded-full bg-white dark:bg-neutral-900 transition-transform ${
                  notifyNewDevice ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="flex items-center justify-between py-1 border-t border-neutral-100 dark:border-neutral-800">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                Device offline
              </div>
              <div className="text-neutral-400 text-[11px]">
                Notify when a monitored device drops off the network
              </div>
            </div>
            <button
              onClick={() => setNotifyDeviceOffline(!notifyDeviceOffline)}
              className={`w-11 h-6 rounded-full transition-colors relative p-0.5 ${
                notifyDeviceOffline ? 'bg-neutral-900 dark:bg-neutral-100' : 'bg-neutral-300 dark:bg-neutral-700'
              }`}
            >
              <div
                className={`w-5 h-5 rounded-full bg-white dark:bg-neutral-900 transition-transform ${
                  notifyDeviceOffline ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="flex items-center justify-between py-1 border-t border-neutral-100 dark:border-neutral-800">
            <div>
              <div className="font-semibold text-neutral-800 dark:text-neutral-200">
                Network changed
              </div>
              <div className="text-neutral-400 text-[11px]">
                Notify when Wi-Fi SSID, default gateway, or IP lease changes
              </div>
            </div>
            <button
              onClick={() => setNotifyNetworkChange(!notifyNetworkChange)}
              className={`w-11 h-6 rounded-full transition-colors relative p-0.5 ${
                notifyNetworkChange ? 'bg-neutral-900 dark:bg-neutral-100' : 'bg-neutral-300 dark:bg-neutral-700'
              }`}
            >
              <div
                className={`w-5 h-5 rounded-full bg-white dark:bg-neutral-900 transition-transform ${
                  notifyNetworkChange ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>
        </div>
      </div>

      {/* DATABASE MAINTENANCE */}
      <div className="bg-white dark:bg-neutral-850 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-3 shadow-xs">
        <div className="text-xs font-bold text-neutral-900 dark:text-neutral-100 uppercase tracking-wider">
          Database Maintenance
        </div>
        <p className="text-xs text-neutral-500">
          Reset local device records, ARP cache, and event history back to initial setup.
        </p>

        {resetMessage && (
          <div className="p-2.5 rounded bg-emerald-50 dark:bg-emerald-950/70 border border-emerald-200 dark:border-emerald-800 text-xs text-emerald-800 dark:text-emerald-300 flex items-center gap-2">
            <Check className="w-3.5 h-3.5 shrink-0" />
            <span>{resetMessage}</span>
          </div>
        )}

        <button
          onClick={handleResetData}
          className="px-3.5 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-red-50 dark:hover:bg-red-950/50 hover:text-red-600 dark:hover:text-red-400 text-neutral-700 dark:text-neutral-300 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 transition-colors flex items-center gap-1.5"
        >
          <Trash2 className="w-3.5 h-3.5" />
          <span>{service.isSimulated ? 'Reset Local Database' : 'Clear History'}</span>
        </button>
        {!service.isSimulated && (
          <button
            onClick={handleOpenData}
            className="ml-2 px-3.5 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-700 dark:text-neutral-300 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 transition-colors inline-flex items-center gap-1.5"
          >
            <FolderOpen className="w-3.5 h-3.5" />
            <span>Open Data Folder</span>
          </button>
        )}
      </div>
    </div>
  );
};
