import React, { useState, useEffect } from 'react';
import { useNetwork } from '../context/NetworkContext';
import { useTheme } from '../context/ThemeContext';
import { AppSettings, DatabaseStats } from '../types/settings';
import {
  ShieldCheck,
  HardDrive,
  Bell,
  Cpu,
  RefreshCw,
  Check,
  Trash2,
  FolderOpen,
  Sun,
  Moon,
  Github,
  ExternalLink,
  Code2,
  Radio,
  Database,
  AlertCircle,
  RotateCw
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

  // Database stats & Maintenance states
  const [dbStats, setDbStats] = useState<DatabaseStats | null>(null);
  const [isBusy, setIsBusy] = useState(false);
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  const showFeedback = (type: 'success' | 'error', message: string) => {
    setFeedback({ type, message });
    setTimeout(() => setFeedback(null), 4000);
  };

  // Load persistent settings & DB stats on mount
  useEffect(() => {
    let isMounted = true;
    if (service.getSettings) {
      service.getSettings().then(s => {
        if (!isMounted || !s) return;
        setAutoDiscovery(s.autoDiscovery);
        setScanInterval(s.scanInterval);
        setNotifyNewDevice(s.notifyNewDevice);
        setNotifyDeviceOffline(s.notifyDeviceOffline);
        setNotifyNetworkChange(s.notifyNetworkChange);
        setLaunchAtStartup(s.launchAtStartup);
        setStartMinimized(s.startMinimized);
      }).catch(() => {});
    }
    if (service.getDatabaseStats) {
      service.getDatabaseStats().then(stats => {
        if (!isMounted || !stats) return;
        setDbStats(stats);
      }).catch(() => {});
    }
    return () => { isMounted = false; };
  }, [service]);

  const saveSetting = async (key: keyof AppSettings, val: any) => {
    try {
      if (service.updateSettings) {
        await service.updateSettings({ [key]: val });
      }
    } catch (err) {
      console.error('Failed to persist setting', key, err);
    }
  };

  const handleResetData = async () => {
    setIsBusy(true);
    try {
      if (service.isSimulated || !service.clearHistory) {
        await setPrototypeStatePreset('normal');
      } else {
        await service.clearHistory();
        await refresh();
      }
      if (service.getDatabaseStats) {
        const stats = await service.getDatabaseStats();
        setDbStats(stats);
      }
      showFeedback('success', 'Local device records and event history were cleared.');
    } catch (e) {
      showFeedback('error', e instanceof Error ? e.message : 'Could not clear history.');
    } finally {
      setIsBusy(false);
    }
  };

  const handleOpenData = async () => {
    try {
      await service.openDataFolder?.();
      showFeedback('success', 'Data folder opened in Windows Explorer.');
    } catch (e) {
      showFeedback('error', e instanceof Error ? e.message : 'Could not open the data folder.');
    }
  };

  const handleVacuum = async () => {
    setIsBusy(true);
    try {
      const res = await service.vacuumDatabase?.();
      if (service.getDatabaseStats) {
        const stats = await service.getDatabaseStats();
        setDbStats(stats);
      }
      showFeedback('success', res?.message || 'Database compacted and defragmented successfully.');
    } catch (e) {
      showFeedback('error', e instanceof Error ? e.message : 'Database VACUUM compaction failed.');
    } finally {
      setIsBusy(false);
    }
  };

  const handleIntegrity = async () => {
    setIsBusy(true);
    try {
      const res = await service.integrityCheck?.();
      showFeedback(res?.success ? 'success' : 'error', res?.message || 'Integrity check finished.');
    } catch (e) {
      showFeedback('error', e instanceof Error ? e.message : 'Database integrity check failed.');
    } finally {
      setIsBusy(false);
    }
  };

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes <= 0) return '0 B';
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
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
      <div className="bg-white dark:bg-neutral-900 border-2 border-emerald-500/30 dark:border-emerald-500/40 rounded-lg p-5 space-y-4 shadow-xs transition-colors">
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
            <span>{dbStats?.dbPath || '%LOCALAPPDATA%\\NetWatch\\data\\network.db'}</span>
            <span className="text-[10px] text-neutral-400 shrink-0 ml-2">{dbStats?.walEnabled !== false ? 'WAL Mode' : 'Rollback Mode'}</span>
          </div>
        </div>
      </div>

      {/* GENERAL (Section 22) */}
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs transition-colors">
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
              onClick={() => {
                const next = !launchAtStartup;
                setLaunchAtStartup(next);
                saveSetting('launchAtStartup', next);
              }}
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
              onClick={() => {
                const next = !startMinimized;
                setStartMinimized(next);
                saveSetting('startMinimized', next);
              }}
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
            <div className="flex gap-1.5 bg-neutral-100 dark:bg-neutral-800/80 p-1 rounded-md border border-neutral-200 dark:border-neutral-700">
              <button
                type="button"
                onClick={() => setTheme('light')}
                className={`px-3 py-1.5 rounded text-xs font-medium transition-all flex items-center gap-1.5 ${
                  theme === 'light'
                    ? 'bg-white text-neutral-900 shadow-xs border border-neutral-200/60 font-semibold'
                    : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-800 dark:hover:text-neutral-200'
                }`}
              >
                <Sun className={`w-3.5 h-3.5 ${theme === 'light' ? 'text-amber-500' : 'text-neutral-400'}`} />
                <span>Light</span>
              </button>
              <button
                type="button"
                onClick={() => setTheme('dark')}
                className={`px-3 py-1.5 rounded text-xs font-medium transition-all flex items-center gap-1.5 ${
                  theme === 'dark'
                    ? 'bg-neutral-700 text-white shadow-xs border border-neutral-600 font-semibold'
                    : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-800 dark:hover:text-neutral-200'
                }`}
              >
                <Moon className={`w-3.5 h-3.5 ${theme === 'dark' ? 'text-sky-300' : 'text-neutral-400'}`} />
                <span>Dark</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* SCANNING SETTINGS (Section 22) */}
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs transition-colors">
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
              onClick={() => {
                const next = !autoDiscovery;
                setAutoDiscovery(next);
                saveSetting('autoDiscovery', next);
              }}
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
              onChange={e => {
                const next = e.target.value;
                setScanInterval(next);
                saveSetting('scanInterval', next);
              }}
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
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs transition-colors">
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
              onClick={() => {
                const next = !notifyNewDevice;
                setNotifyNewDevice(next);
                saveSetting('notifyNewDevice', next);
              }}
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
              onClick={() => {
                const next = !notifyDeviceOffline;
                setNotifyDeviceOffline(next);
                saveSetting('notifyDeviceOffline', next);
              }}
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
              onClick={() => {
                const next = !notifyNetworkChange;
                setNotifyNetworkChange(next);
                saveSetting('notifyNetworkChange', next);
              }}
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

      {/* DATABASE MAINTENANCE & LIVE INSPECTION */}
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs transition-colors">
        <div className="flex items-center justify-between flex-wrap gap-2">
          <div>
            <div className="text-xs font-bold text-neutral-900 dark:text-neutral-100 uppercase tracking-wider flex items-center gap-2">
              <Database className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
              <span>Database Inspection & Maintenance</span>
            </div>
            <p className="text-xs text-neutral-500 dark:text-neutral-400 mt-0.5">
              Monitor local SQLite health, execute storage maintenance, and inspect data files.
            </p>
          </div>
          {dbStats && (
            <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
              {dbStats.walEnabled ? 'WAL Mode Active' : 'Rollback Journal'}
            </span>
          )}
        </div>

        {/* Live DB Stats Card */}
        <div className="p-3 rounded bg-neutral-50 dark:bg-neutral-950/60 border border-neutral-200 dark:border-neutral-800 grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          <div>
            <span className="text-[10px] uppercase font-semibold text-neutral-400 block">Database Size</span>
            <span className="font-mono font-medium text-neutral-800 dark:text-neutral-200">
              {formatBytes(dbStats?.fileSizeBytes ?? 61440)}
            </span>
          </div>
          <div>
            <span className="text-[10px] uppercase font-semibold text-neutral-400 block">Devices Stored</span>
            <span className="font-mono font-medium text-neutral-800 dark:text-neutral-200">
              {dbStats?.deviceCount ?? 0}
            </span>
          </div>
          <div>
            <span className="text-[10px] uppercase font-semibold text-neutral-400 block">Event History</span>
            <span className="font-mono font-medium text-neutral-800 dark:text-neutral-200">
              {dbStats?.eventCount ?? 0}
            </span>
          </div>
          <div>
            <span className="text-[10px] uppercase font-semibold text-neutral-400 block">Scan Records</span>
            <span className="font-mono font-medium text-neutral-800 dark:text-neutral-200">
              {dbStats?.scanCount ?? 0}
            </span>
          </div>
        </div>

        {/* Dynamic File Path */}
        <div className="p-2.5 rounded bg-neutral-100 dark:bg-neutral-900 font-mono text-[11px] text-neutral-700 dark:text-neutral-300 flex items-center justify-between break-all border border-neutral-200 dark:border-neutral-800">
          <span className="truncate">{dbStats?.dbPath || '%LOCALAPPDATA%\\NetWatch\\data\\network.db'}</span>
          <span className="text-[10px] text-emerald-600 dark:text-emerald-400 shrink-0 ml-2 font-sans font-medium">Local SQLite</span>
        </div>

        {/* Feedback alert (distinct success vs error) */}
        {feedback && (
          <div className={`p-2.5 rounded text-xs flex items-center gap-2 border ${
            feedback.type === 'success'
              ? 'bg-emerald-50 dark:bg-emerald-950/70 border-emerald-200 dark:border-emerald-800 text-emerald-800 dark:text-emerald-300'
              : 'bg-red-50 dark:bg-red-950/70 border-red-200 dark:border-red-800 text-red-800 dark:text-red-300'
          }`}>
            {feedback.type === 'success' ? (
              <Check className="w-3.5 h-3.5 shrink-0" />
            ) : (
              <AlertCircle className="w-3.5 h-3.5 shrink-0" />
            )}
            <span>{feedback.message}</span>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex flex-wrap items-center gap-2 pt-1">
          <button
            onClick={handleOpenData}
            disabled={isBusy}
            className="px-3.5 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-700 dark:text-neutral-300 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 transition-colors inline-flex items-center gap-1.5 disabled:opacity-50"
          >
            <FolderOpen className="w-3.5 h-3.5" />
            <span>Open Data Folder</span>
          </button>

          <button
            onClick={handleVacuum}
            disabled={isBusy}
            className="px-3.5 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-700 dark:text-neutral-300 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 transition-colors inline-flex items-center gap-1.5 disabled:opacity-50"
            title="Defragment and shrink database file on disk"
          >
            <RotateCw className={`w-3.5 h-3.5 ${isBusy ? 'animate-spin' : ''}`} />
            <span>Compact (VACUUM)</span>
          </button>

          <button
            onClick={handleIntegrity}
            disabled={isBusy}
            className="px-3.5 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-700 dark:text-neutral-300 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 transition-colors inline-flex items-center gap-1.5 disabled:opacity-50"
            title="Execute SQLite PRAGMA integrity_check"
          >
            <ShieldCheck className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
            <span>Check Integrity</span>
          </button>

          <button
            onClick={handleResetData}
            disabled={isBusy}
            className="px-3.5 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-red-50 dark:hover:bg-red-950/50 hover:text-red-600 dark:hover:text-red-400 text-neutral-700 dark:text-neutral-300 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 transition-colors flex items-center gap-1.5 disabled:opacity-50"
          >
            <Trash2 className="w-3.5 h-3.5" />
            <span>{service.isSimulated ? 'Reset Mock Data' : 'Clear History'}</span>
          </button>
        </div>
      </div>

      {/* About & Developer Identity */}
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 shadow-xs transition-colors">
        <div className="flex items-start justify-between flex-wrap gap-3">
          <div className="flex items-center gap-3">
            <img src="/favicon.svg" alt="NETWATCH Logo" className="w-10 h-10 rounded-xl shadow-xs shrink-0" />
            <div>
              <h2 className="text-sm font-semibold text-neutral-900 dark:text-neutral-100 flex items-center gap-2">
                NETWATCH Desktop
                <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                  v0.1.0-dev
                </span>
              </h2>
              <p className="text-xs text-neutral-500 dark:text-neutral-400 mt-0.5">
                Super Fast Network Scanner &amp; Local Hardware Inventory
              </p>
            </div>
          </div>

          <a
            href="https://github.com/alwkala/NETWATCH"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-700 dark:text-neutral-200 border border-neutral-300 dark:border-neutral-700 transition-colors"
          >
            <Github className="w-3.5 h-3.5" />
            <span>GitHub Repository</span>
            <ExternalLink className="w-3 h-3 text-neutral-400" />
          </a>
        </div>

        <div className="mt-4 pt-4 border-t border-neutral-100 dark:border-neutral-800/80 grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
          <div className="p-3 rounded bg-neutral-50 dark:bg-neutral-950/50 border border-neutral-200/60 dark:border-neutral-800/60">
            <span className="text-[10px] uppercase tracking-wider font-semibold text-neutral-400 block mb-1">
              Architecture
            </span>
            <span className="font-medium text-neutral-700 dark:text-neutral-300 flex items-center gap-1.5">
              <Cpu className="w-3.5 h-3.5 text-emerald-500" />
              Go Engine + Wails v2
            </span>
          </div>

          <div className="p-3 rounded bg-neutral-50 dark:bg-neutral-950/50 border border-neutral-200/60 dark:border-neutral-800/60">
            <span className="text-[10px] uppercase tracking-wider font-semibold text-neutral-400 block mb-1">
              Data Privacy
            </span>
            <span className="font-medium text-emerald-600 dark:text-emerald-400 flex items-center gap-1.5">
              <ShieldCheck className="w-3.5 h-3.5" />
              100% Offline / Zero Telemetry
            </span>
          </div>

          <div className="p-3 rounded bg-neutral-50 dark:bg-neutral-950/50 border border-neutral-200/60 dark:border-neutral-800/60">
            <span className="text-[10px] uppercase tracking-wider font-semibold text-neutral-400 block mb-1">
              Studio &amp; Engineering
            </span>
            <span className="font-medium text-neutral-700 dark:text-neutral-300 flex items-center gap-1.5">
              <Code2 className="w-3.5 h-3.5 text-blue-500" />
              Engineered by Alwkala
            </span>
          </div>
        </div>

        <div className="mt-3 text-[11px] text-neutral-400 dark:text-neutral-500 flex items-center justify-between flex-wrap gap-2">
          <span>Dual Licensed under MIT &amp; Apache 2.0</span>
          <span>Know Every Device on Your LAN. Without the Cloud Watching.</span>
        </div>
      </div>
    </div>
  );
};
