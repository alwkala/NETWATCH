import React from 'react';
import { useNetwork, AppPage } from '../../context/NetworkContext';
import { useTheme } from '../../context/ThemeContext';
import {
  RotateCw,
  Radar,
  Wifi,
  ChevronRight,
  Sun,
  Moon,
  Menu
} from 'lucide-react';

interface TopBarProps {
  onToggleMobileMenu?: () => void;
}

export const TopBar: React.FC<TopBarProps> = ({ onToggleMobileMenu }) => {
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
  const { theme, toggleTheme } = useTheme();

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
    <div className="h-13 shrink-0 bg-white dark:bg-neutral-900 border-b border-neutral-200 dark:border-neutral-800 px-3 md:px-6 flex items-center justify-between transition-colors">
      {/* Left: Mobile menu toggle + Breadcrumb / Title */}
      <div className="flex items-center gap-2 min-w-0">
        <button
          onClick={onToggleMobileMenu}
          className="p-1.5 md:hidden text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded border border-neutral-200 dark:border-neutral-700 transition-colors shrink-0"
          title="Open navigation menu"
          aria-label="Open navigation menu"
        >
          <Menu className="w-4 h-4" />
        </button>

        <h1 className="text-sm md:text-base font-bold text-neutral-900 dark:text-neutral-100 tracking-tight truncate">
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
        <div className="hidden md:flex items-center gap-2 text-xs text-neutral-600 dark:text-neutral-400 px-3 py-1.5 rounded bg-neutral-100 dark:bg-neutral-800/80 border border-neutral-200 dark:border-neutral-700 font-mono">
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

        {/* Theme Toggle Button */}
        <button
          onClick={toggleTheme}
          className="p-1.5 text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded border border-neutral-200 dark:border-neutral-700 transition-colors"
          title={`Switch to ${theme === 'dark' ? 'Light' : 'Dark'} mode`}
          aria-label="Toggle color theme"
        >
          {theme === 'dark' ? (
            <Sun className="w-4 h-4 text-amber-400" />
          ) : (
            <Moon className="w-4 h-4 text-neutral-600" />
          )}
        </button>

        {/* Global Scan Action (Hidden when already on dedicated Scanner page) */}
        {activePage !== 'scanner' && (
          <button
            onClick={() => {
              navigateTo('scanner');
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
                <Radar className="w-3.5 h-3.5 text-emerald-500 dark:text-emerald-600" />
                <span>Scan</span>
              </>
            )}
          </button>
        )}
      </div>
    </div>
  );
};
