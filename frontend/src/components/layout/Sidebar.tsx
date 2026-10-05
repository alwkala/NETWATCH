import React from 'react';
import { useNetwork, AppPage } from '../../context/NetworkContext';
import {
  LayoutDashboard,
  HardDrive,
  Network,
  Radar,
  Clock,
  Settings,
  ShieldCheck,
  Wifi,
  Radio
} from 'lucide-react';

export const Sidebar: React.FC = () => {
  const {
    activePage,
    navigateTo,
    devices,
    events,
    networkInfo,
    isScanning
  } = useNetwork();

  const onlineDevicesCount = devices.filter(d => d.status === 'online').length;
  const newDevicesCount = devices.filter(d => d.isNew).length;

  const navItemClass = (page: AppPage) => {
    const isActive = activePage === page;
    return `w-full flex items-center justify-between px-3 py-2 rounded-md text-xs font-medium transition-colors ${
      isActive
        ? 'bg-neutral-200/80 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 font-semibold'
        : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-150 dark:hover:bg-neutral-850 hover:text-neutral-900 dark:hover:text-neutral-200'
    }`;
  };

  return (
    <aside className="w-56 shrink-0 bg-neutral-50 dark:bg-neutral-900/90 border-r border-neutral-200 dark:border-neutral-800 flex flex-col justify-between select-none">
      {/* Top Header & Navigation */}
      <div className="p-3">
        {/* Logo Lockup */}
        <div className="flex items-center gap-2.5 px-2 py-2 mb-4">
          <div className="w-6 h-6 rounded bg-neutral-900 dark:bg-neutral-100 flex items-center justify-center text-white dark:text-neutral-900 shadow-xs">
            <Radio className="w-3.5 h-3.5 text-emerald-500 dark:text-emerald-600" />
          </div>
          <div>
            <div className="text-sm font-bold tracking-tight text-neutral-900 dark:text-neutral-100 leading-none">
              NETWATCH
            </div>
            <div className="text-[10px] text-neutral-500 dark:text-neutral-400 leading-none mt-1">
              Desktop Edition
            </div>
          </div>
        </div>

        {/* Navigation Categories */}
        <nav className="space-y-4">
          {/* OVERVIEW */}
          <div>
            <div className="px-2 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider mb-1">
              Overview
            </div>
            <button
              onClick={() => navigateTo('dashboard')}
              className={navItemClass('dashboard')}
            >
              <div className="flex items-center gap-2.5">
                <LayoutDashboard className="w-4 h-4" />
                <span>Dashboard</span>
              </div>
            </button>
          </div>

          {/* NETWORK */}
          <div>
            <div className="px-2 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider mb-1">
              Network
            </div>
            <div className="space-y-0.5">
              <button
                onClick={() => navigateTo('devices')}
                className={navItemClass('devices')}
              >
                <div className="flex items-center gap-2.5">
                  <HardDrive className="w-4 h-4" />
                  <span>Devices</span>
                </div>
                <div className="flex items-center gap-1 font-mono text-[11px] tabular-nums">
                  <span className="text-neutral-700 dark:text-neutral-300 font-semibold">
                    {onlineDevicesCount}
                  </span>
                  <span className="text-neutral-400 dark:text-neutral-500">
                    /{devices.length}
                  </span>
                  {newDevicesCount > 0 && (
                    <span className="w-1.5 h-1.5 rounded-full bg-amber-500 ml-0.5" title="New device discovered" />
                  )}
                </div>
              </button>

              <button
                onClick={() => navigateTo('network')}
                className={navItemClass('network')}
              >
                <div className="flex items-center gap-2.5">
                  <Network className="w-4 h-4" />
                  <span>Network</span>
                </div>
              </button>

              <button
                onClick={() => navigateTo('scanner')}
                className={navItemClass('scanner')}
              >
                <div className="flex items-center gap-2.5">
                  <Radar className={`w-4 h-4 ${isScanning ? 'animate-spin text-emerald-500' : ''}`} />
                  <span>Scanner</span>
                </div>
                {isScanning && (
                  <span className="text-[10px] text-emerald-600 dark:text-emerald-400 font-medium">
                    ACTIVE
                  </span>
                )}
              </button>
            </div>
          </div>

          {/* ACTIVITY */}
          <div>
            <div className="px-2 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider mb-1">
              Activity
            </div>
            <button
              onClick={() => navigateTo('events')}
              className={navItemClass('events')}
            >
              <div className="flex items-center gap-2.5">
                <Clock className="w-4 h-4" />
                <span>Events</span>
              </div>
              <span className="text-[11px] font-mono text-neutral-500 dark:text-neutral-400 tabular-nums">
                {events.length}
              </span>
            </button>
          </div>
        </nav>
      </div>

      {/* Bottom Area: Privacy Badge, Settings, Interface Status */}
      <div className="p-3 border-t border-neutral-200 dark:border-neutral-800 space-y-2">
        {/* Privacy Assurance Box */}
        <div className="px-2.5 py-2 rounded bg-neutral-100 dark:bg-neutral-850 border border-neutral-200/80 dark:border-neutral-800 text-[11px]">
          <div className="flex items-center gap-1.5 text-neutral-800 dark:text-neutral-200 font-semibold mb-0.5">
            <ShieldCheck className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
            <span>Privacy Guaranteed</span>
          </div>
          <p className="text-[10px] text-neutral-500 dark:text-neutral-400 leading-tight">
            100% on-device SQLite. No cloud, telemetry, or external analytics.
          </p>
        </div>

        {/* Settings button */}
        <button
          onClick={() => navigateTo('settings')}
          className={navItemClass('settings')}
        >
          <div className="flex items-center gap-2.5">
            <Settings className="w-4 h-4" />
            <span>Settings & Privacy</span>
          </div>
        </button>

        {/* Active Interface Status */}
        <div className="px-2 pt-1 flex items-center justify-between text-[11px] text-neutral-500 dark:text-neutral-400 font-mono">
          <div className="flex items-center gap-1.5 truncate">
            <Wifi className="w-3 h-3 text-emerald-500 shrink-0" />
            <span className="truncate">{networkInfo?.interfaceName || 'Wi-Fi'}</span>
          </div>
          <span className="tabular-nums shrink-0">{networkInfo?.localIp || '192.168.1.24'}</span>
        </div>
      </div>
    </aside>
  );
};
