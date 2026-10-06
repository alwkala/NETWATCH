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
  Github,
  X
} from 'lucide-react';

interface SidebarProps {
  mobileOpen?: boolean;
  onMobileClose?: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ mobileOpen = false, onMobileClose }) => {
  const {
    activePage,
    navigateTo,
    devices,
    events,
    networkInfo
  } = useNetwork();

  const onlineDevicesCount = devices.filter(d => d.status === 'online').length;
  const newDevicesCount = devices.filter(d => d.isNew).length;

  const handleNav = (page: AppPage) => {
    navigateTo(page);
    if (onMobileClose) {
      onMobileClose();
    }
  };

  const navItemClass = (page: AppPage) => {
    const isActive = activePage === page;
    return `w-full flex items-center justify-between px-3 py-2 rounded-md text-xs font-medium transition-colors ${
      isActive
        ? 'bg-neutral-200/80 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 font-semibold'
        : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-200/60 dark:hover:bg-neutral-800 hover:text-neutral-900 dark:hover:text-neutral-200'
    }`;
  };

  const renderContent = (isMobile = false) => (
    <>
      {/* Top Header & Navigation */}
      <div className="p-3">
        {/* Logo Lockup - STRICT INVARIANT: Header text must remain untouched */}
        <div className="flex items-center justify-between px-2 py-2 mb-4">
          <div className="flex items-center gap-2.5">
            <img src="/favicon.svg" alt="NETWATCH" className="w-6 h-6 rounded shadow-xs shrink-0" />
            <div>
              <div className="text-sm font-bold tracking-tight text-neutral-900 dark:text-neutral-100 leading-none">
                NETWATCH
              </div>
              <div className="text-[10px] text-neutral-500 dark:text-neutral-400 leading-none mt-1">
                Desktop Edition
              </div>
            </div>
          </div>
          {isMobile && onMobileClose && (
            <button
              onClick={onMobileClose}
              className="p-1 rounded text-neutral-400 hover:text-neutral-700 dark:hover:text-neutral-200 hover:bg-neutral-200 dark:hover:bg-neutral-800 transition-colors"
              title="Close navigation"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Navigation Categories */}
        <nav className="space-y-4">
          {/* OVERVIEW */}
          <div>
            <div className="px-2 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider mb-1">
              Overview
            </div>
            <button
              onClick={() => handleNav('dashboard')}
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
                onClick={() => handleNav('devices')}
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
                onClick={() => handleNav('network')}
                className={navItemClass('network')}
              >
                <div className="flex items-center gap-2.5">
                  <Network className="w-4 h-4" />
                  <span>Network</span>
                </div>
              </button>

              <button
                onClick={() => handleNav('scanner')}
                className={navItemClass('scanner')}
              >
                <div className="flex items-center gap-2.5">
                  <Radar className="w-4 h-4" />
                  <span>Scanner</span>
                </div>
              </button>
            </div>
          </div>

          {/* ACTIVITY */}
          <div>
            <div className="px-2 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider mb-1">
              Activity
            </div>
            <button
              onClick={() => handleNav('events')}
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
        <div className="px-2.5 py-2 rounded bg-neutral-100 dark:bg-neutral-800/70 border border-neutral-200/80 dark:border-neutral-700/60 text-[11px]">
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
          onClick={() => handleNav('settings')}
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

        {/* Developer & Studio Identity Capsule */}
        <div className="pt-2 border-t border-neutral-200/80 dark:border-neutral-800">
          <div className="px-2.5 py-1.5 rounded-md bg-neutral-200/60 dark:bg-neutral-800/60 border border-neutral-300/50 dark:border-neutral-700/70 flex items-center justify-between">
            <div className="flex items-center gap-1.5 min-w-0">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shrink-0" title="Engine Active" />
              <div className="min-w-0">
                <div className="text-[10px] font-bold text-neutral-800 dark:text-neutral-200 truncate leading-tight flex items-center gap-1">
                  <span>NETWATCH</span>
                  <span className="font-mono text-[9px] font-normal text-neutral-500 dark:text-neutral-400">v0.2.0-alpha.1</span>
                </div>
                <div className="text-[9px] text-neutral-500 dark:text-neutral-400 truncate leading-tight">
                  By <span className="font-semibold text-neutral-700 dark:text-neutral-300">Alwkala</span>
                </div>
              </div>
            </div>
            <a
              href="https://github.com/alwkala/NETWATCH"
              target="_blank"
              rel="noopener noreferrer"
              className="p-1 text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-300/50 dark:hover:bg-neutral-700/50 rounded transition-colors"
              title="View source on GitHub (alwkala/NETWATCH)"
            >
              <Github className="w-3.5 h-3.5" />
            </a>
          </div>
        </div>
      </div>
    </>
  );

  return (
    <>
      {/* Desktop Persistent Sidebar */}
      <aside className="hidden md:flex w-56 shrink-0 bg-neutral-50 dark:bg-neutral-900/90 border-r border-neutral-200 dark:border-neutral-800 flex-col justify-between select-none">
        {renderContent(false)}
      </aside>

      {/* Mobile Backdrop Overlay */}
      {mobileOpen && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-xs z-40 md:hidden transition-opacity"
          onClick={onMobileClose}
          aria-hidden="true"
        />
      )}

      {/* Mobile Slide-Over Drawer */}
      <aside
        className={`fixed inset-y-0 left-0 w-64 bg-neutral-50 dark:bg-neutral-900 border-r border-neutral-200 dark:border-neutral-800 flex flex-col justify-between select-none z-50 transition-transform duration-200 ease-in-out md:hidden shadow-2xl ${
          mobileOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {renderContent(true)}
      </aside>
    </>
  );
};
