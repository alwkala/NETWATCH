import React from 'react';
import { useNetwork, AppPage } from '../../context/NetworkContext';
import {
  RotateCw,
  Search,
  Wifi,
  ChevronRight,
  Sparkles
} from 'lucide-react';

export const TopBar: React.FC = () => {
  const {
    activePage,
    selectedDevice,
    networkInfo,
    isScanning,
    scanProgress,
    startScan,
    refresh,
    navigateTo
  } = useNetwork();

  const getPageTitle = () => {
    switch (activePage) {
      case 'dashboard':
        return 'Network Overview';
      case 'devices':
        return 'Devices Inventory';
      case 'device-details':
        return 'Device Inspector';
      case 'network':
        return 'Subnet & Interfaces';
      case 'scanner':
        return 'Network Scanner';
      case 'events':
        return 'Activity History';
      case 'settings':
        return 'Settings & Privacy';
      default:
        return 'NetWatch';
    }
  };

  return (
    <div className="h-13 shrink-0 bg-white dark:bg-neutral-850 border-b border-neutral-200 dark:border-neutral-800 px-6 flex items-center justify-between">
      {/* Left: Breadcrumb / Title */}
      <div className="flex items-center gap-2">
        <h1 className="text-base font-bold text-neutral-900 dark:text-neutral-100 tracking-tight">
          {getPageTitle()}
        </h1>
        {activePage === 'device-details' && selectedDevice && (
          <div className="flex items-center gap-1.5 text-xs text-neutral-400">
            <ChevronRight className="w-3.5 h-3.5" />
            <button
              onClick={() => navigateTo('devices')}
              className="hover:underline hover:text-neutral-700 dark:hover:text-neutral-200"
            >
              Devices
            </button>
            <ChevronRight className="w-3.5 h-3.5" />
            <span className="font-mono text-neutral-700 dark:text-neutral-300 font-medium">
              {selectedDevice.name}
            </span>
          </div>
        )}
      </div>

      {/* Right: Current network indicator & primary action */}
      <div className="flex items-center gap-3">
        {/* Network & Subnet indicator */}
        <div className="hidden md:flex items-center gap-2 text-xs text-neutral-600 dark:text-neutral-400 px-3 py-1.5 rounded bg-neutral-100 dark:bg-neutral-800/80 border border-neutral-200 dark:border-neutral-750 font-mono">
          <Wifi className="w-3.5 h-3.5 text-neutral-500" />
          <span className="font-sans font-medium text-neutral-800 dark:text-neutral-200">
            {networkInfo?.ssid || networkInfo?.interfaceName || 'No network'}
          </span>
          <span className="text-neutral-400 dark:text-neutral-600">·</span>
          <span>{networkInfo?.subnet || '192.168.1.0/24'}</span>
          <span className="text-neutral-400 dark:text-neutral-600">·</span>
          <span className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400 font-sans font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            Connected
          </span>
        </div>

        {/* Refresh button */}
        <button
          onClick={() => refresh()}
          disabled={isScanning}
          className="p-1.5 text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded border border-neutral-200 dark:border-neutral-700 transition-colors disabled:opacity-50"
          title="Refresh ARP/NDP cache"
          aria-label="Refresh network data"
        >
          <RotateCw className="w-4 h-4" />
        </button>

        {/* Primary Scan Action */}
        <button
          onClick={() => {
            if (activePage !== 'scanner') {
              navigateTo('scanner');
            }
            startScan('quick');
          }}
          disabled={isScanning}
          className="px-3.5 py-1.5 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-950 hover:bg-neutral-800 dark:hover:bg-white text-xs font-semibold rounded transition-colors shadow-xs flex items-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {isScanning ? (
            <>
              <RotateCw className="w-3.5 h-3.5 animate-spin" />
              <span>Scanning... {scanProgress.progress}%</span>
            </>
          ) : (
            <>
              <Search className="w-3.5 h-3.5" />
              <span>Scan Network</span>
            </>
          )}
        </button>
      </div>
    </div>
  );
};
