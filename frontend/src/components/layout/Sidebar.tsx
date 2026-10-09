import React from 'react';
import { useNetwork, AppPage } from '../../context/NetworkContext';
import { useI18n } from '../../context/I18nContext';
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
  X,
  PanelLeftClose,
  PanelLeft
} from 'lucide-react';

interface SidebarProps {
  mobileOpen?: boolean;
  onMobileClose?: () => void;
  collapsed?: boolean;
  onToggleCollapse?: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  mobileOpen = false,
  onMobileClose,
  collapsed = false,
  onToggleCollapse
}) => {
  const {
    activePage,
    navigateTo,
    devices,
    events,
    networkInfo
  } = useNetwork();
  const { t, isRTL } = useI18n();

  const onlineDevicesCount = devices.filter(d => d.status === 'online').length;
  const newDevicesCount = devices.filter(d => d.isNew).length;

  const handleNav = (page: AppPage) => {
    navigateTo(page);
    if (onMobileClose) {
      onMobileClose();
    }
  };

  const navItemClass = (page: AppPage, isCollapsed = false) => {
    const isActive = activePage === page;
    return `w-full flex items-center ${isCollapsed ? 'justify-center p-2.5' : 'justify-between px-3 py-2'} rounded-md text-xs font-medium transition-colors cursor-pointer ${
      isActive
        ? 'bg-neutral-200/80 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 font-semibold'
        : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-200/60 dark:hover:bg-neutral-800 hover:text-neutral-900 dark:hover:text-neutral-200'
    }`;
  };

  const renderContent = (isMobile = false, isCollapsed = false) => (
    <>
      {/* Top Header & Navigation */}
      <div className={isCollapsed ? 'p-2' : 'p-3'}>
        {/* Logo Lockup - STRICT INVARIANT: Header text must remain untouched */}
        <div className={`flex items-center ${isCollapsed ? 'justify-center' : 'justify-between'} px-1 py-2 mb-4`}>
          <div className="flex items-center gap-2.5 min-w-0">
            <img src="/favicon.svg" alt="NETWATCH" className="w-6 h-6 rounded shadow-xs shrink-0" />
            {!isCollapsed && (
              <div className="min-w-0">
                <div className="text-sm font-bold tracking-tight text-neutral-900 dark:text-neutral-100 leading-none">
                  NETWATCH
                </div>
                <div className="text-[10px] text-neutral-500 dark:text-neutral-400 leading-none mt-1">
                  Desktop Edition
                </div>
              </div>
            )}
          </div>
          {!isMobile && onToggleCollapse && (
            <button
              onClick={onToggleCollapse}
              className="p-1 rounded text-neutral-400 hover:text-neutral-700 dark:hover:text-neutral-200 hover:bg-neutral-200 dark:hover:bg-neutral-800 transition-colors cursor-pointer"
              title={isCollapsed ? (t('nav.expandSidebar') || 'Expand') : (t('nav.collapseSidebar') || 'Collapse')}
              aria-label="Toggle sidebar"
            >
              {isCollapsed ? <PanelLeft className="w-4 h-4" /> : <PanelLeftClose className="w-4 h-4" />}
            </button>
          )}
          {isMobile && onMobileClose && (
            <button
              onClick={onMobileClose}
              className="p-1 rounded text-neutral-400 hover:text-neutral-700 dark:hover:text-neutral-200 hover:bg-neutral-200 dark:hover:bg-neutral-800 transition-colors cursor-pointer"
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
            {!isCollapsed && (
              <div className="px-2 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider mb-1">
                {t('nav.overview')}
              </div>
            )}
            <button
              onClick={() => handleNav('dashboard')}
              className={navItemClass('dashboard', isCollapsed)}
              title={isCollapsed ? t('nav.dashboard') : undefined}
            >
              <div className={`flex items-center ${isCollapsed ? 'justify-center' : 'gap-2.5'}`}>
                <LayoutDashboard className="w-4 h-4" />
                {!isCollapsed && <span>{t('nav.dashboard')}</span>}
              </div>
            </button>
          </div>

          {/* NETWORK */}
          <div>
            {!isCollapsed && (
              <div className="px-2 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider mb-1">
                {t('nav.networkSection')}
              </div>
            )}
            <div className="space-y-0.5">
              <button
                onClick={() => handleNav('devices')}
                className={navItemClass('devices', isCollapsed)}
                title={isCollapsed ? `${t('nav.devices')} (${onlineDevicesCount}/${devices.length})` : undefined}
              >
                <div className={`flex items-center ${isCollapsed ? 'justify-center relative' : 'gap-2.5'}`}>
                  <HardDrive className="w-4 h-4" />
                  {!isCollapsed && <span>{t('nav.devices')}</span>}
                  {isCollapsed && newDevicesCount > 0 && (
                    <span className="w-2 h-2 rounded-full bg-amber-500 absolute -top-1 -right-1" />
                  )}
                </div>
                {!isCollapsed && (
                  <div className="flex items-center gap-1 font-mono text-[11px] tabular-nums" dir="ltr">
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
                )}
              </button>

              <button
                onClick={() => handleNav('network')}
                className={navItemClass('network', isCollapsed)}
                title={isCollapsed ? t('nav.network') : undefined}
              >
                <div className={`flex items-center ${isCollapsed ? 'justify-center' : 'gap-2.5'}`}>
                  <Network className="w-4 h-4" />
                  {!isCollapsed && <span>{t('nav.network')}</span>}
                </div>
              </button>

              <button
                onClick={() => handleNav('scanner')}
                className={navItemClass('scanner', isCollapsed)}
                title={isCollapsed ? t('nav.scanner') : undefined}
              >
                <div className={`flex items-center ${isCollapsed ? 'justify-center' : 'gap-2.5'}`}>
                  <Radar className="w-4 h-4" />
                  {!isCollapsed && <span>{t('nav.scanner')}</span>}
                </div>
              </button>
            </div>
          </div>

          {/* ACTIVITY */}
          <div>
            {!isCollapsed && (
              <div className="px-2 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider mb-1">
                {t('nav.activitySection')}
              </div>
            )}
            <button
              onClick={() => handleNav('events')}
              className={navItemClass('events', isCollapsed)}
              title={isCollapsed ? `${t('nav.events')} (${events.length})` : undefined}
            >
              <div className={`flex items-center ${isCollapsed ? 'justify-center' : 'gap-2.5'}`}>
                <Clock className="w-4 h-4" />
                {!isCollapsed && <span>{t('nav.events')}</span>}
              </div>
              {!isCollapsed && (
                <span className="text-[11px] font-mono text-neutral-500 dark:text-neutral-400 tabular-nums">
                  {events.length}
                </span>
              )}
            </button>
          </div>
        </nav>
      </div>

      {/* Bottom Area: Settings -> Network Status -> Privacy Badge -> Developer Capsule */}
      <div className={`${isCollapsed ? 'p-2' : 'p-3'} border-t border-neutral-200 dark:border-neutral-800 space-y-2`}>
        {/* 1. Settings button - At the top of the bottom section */}
        <button
          onClick={() => handleNav('settings')}
          className={navItemClass('settings', isCollapsed)}
          title={isCollapsed ? t('nav.settings') : undefined}
        >
          <div className={`flex items-center ${isCollapsed ? 'justify-center' : 'gap-2.5'}`}>
            <Settings className="w-4 h-4" />
            {!isCollapsed && <span>{t('nav.settings')}</span>}
          </div>
        </button>

        {/* 2. Active Interface Status (اشعار الشبكة) */}
        {!isCollapsed ? (
          <div className="px-2.5 py-1.5 rounded bg-neutral-100 dark:bg-neutral-800/50 border border-neutral-200/60 dark:border-neutral-700/60 flex items-center justify-between text-[11px] text-neutral-600 dark:text-neutral-400 font-mono">
            <div className="flex items-center gap-1.5 truncate">
              <Wifi className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
              <span className="truncate font-sans font-medium text-neutral-800 dark:text-neutral-200">
                {networkInfo?.interfaceName || 'Ethernet'}
              </span>
            </div>
            <span className="tabular-nums shrink-0 text-[10px]" dir="ltr">
              {networkInfo?.localIp || '192.168.1.1'}
            </span>
          </div>
        ) : (
          <div
            className="p-2 rounded bg-neutral-100 dark:bg-neutral-800/50 flex items-center justify-center text-emerald-500"
            title={`${networkInfo?.interfaceName || 'Ethernet'} - ${networkInfo?.localIp || '192.168.1.1'}`}
          >
            <Wifi className="w-4 h-4" />
          </div>
        )}

        {/* 3. Privacy Assurance Box - Compact Icon + Title + Single Subtitle */}
        {!isCollapsed ? (
          <div className="px-2.5 py-1.5 rounded bg-neutral-100 dark:bg-neutral-800/50 border border-neutral-200/60 dark:border-neutral-700/60 flex items-center gap-2">
            <ShieldCheck className="w-4 h-4 text-emerald-600 dark:text-emerald-400 shrink-0" />
            <div className="min-w-0 flex-1">
              <div className="text-[11px] font-semibold text-neutral-800 dark:text-neutral-200 leading-tight truncate">
                {t('nav.privacyGuaranteed')}
              </div>
              <div className="text-[10px] text-neutral-500 dark:text-neutral-400 leading-tight truncate">
                {t('nav.privacySubtitle')}
              </div>
            </div>
          </div>
        ) : (
          <div
            className="p-2 rounded bg-neutral-100 dark:bg-neutral-800/50 flex items-center justify-center text-emerald-600 dark:text-emerald-400"
            title={`${t('nav.privacyGuaranteed')} - ${t('nav.privacySubtitle')}`}
          >
            <ShieldCheck className="w-4 h-4" />
          </div>
        )}

        {/* 4. Developer & Studio Identity Capsule */}
        <div className="pt-1.5 border-t border-neutral-200/80 dark:border-neutral-800">
          {!isCollapsed ? (
            <div className="px-2.5 py-1.5 rounded-md bg-neutral-200/60 dark:bg-neutral-800/60 border border-neutral-300/50 dark:border-neutral-700/70 flex items-center justify-between">
              <div className="flex items-center gap-1.5 min-w-0">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shrink-0" title="Engine Active" />
                <div className="min-w-0">
                  <div className="text-[10px] font-bold text-neutral-800 dark:text-neutral-200 truncate leading-tight flex items-center gap-1">
                    <span>NETWATCH</span>
                    <span className="font-mono text-[9px] font-normal text-neutral-500 dark:text-neutral-400">v0.3.0-alpha.1</span>
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
          ) : (
            <div className="flex items-center justify-center py-0.5">
              <a
                href="https://github.com/alwkala/NETWATCH"
                target="_blank"
                rel="noopener noreferrer"
                className="p-1.5 text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-200 dark:hover:bg-neutral-800 rounded transition-colors"
                title="NETWATCH v0.3.0-alpha.1 By Alwkala"
              >
                <Github className="w-4 h-4" />
              </a>
            </div>
          )}
        </div>
      </div>
    </>
  );

  return (
    <>
      {/* Desktop Persistent Sidebar */}
      <aside
        className={`hidden md:flex ${
          collapsed ? 'w-16' : 'w-56'
        } shrink-0 bg-neutral-50 dark:bg-neutral-900/90 border-r rtl:border-r-0 rtl:border-l border-neutral-200 dark:border-neutral-800 flex-col justify-between select-none transition-[width] duration-200 ease-in-out`}
      >
        {renderContent(false, collapsed)}
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
        className={`fixed inset-y-0 inset-s-0 w-64 bg-neutral-50 dark:bg-neutral-900 border-r rtl:border-r-0 rtl:border-l border-neutral-200 dark:border-neutral-800 flex flex-col justify-between select-none z-50 transition-transform duration-200 ease-in-out md:hidden shadow-2xl ${
          mobileOpen ? 'translate-x-0' : (isRTL ? 'translate-x-full' : '-translate-x-full')
        }`}
      >
        {renderContent(true, false)}
      </aside>
    </>
  );
};
