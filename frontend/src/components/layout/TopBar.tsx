import React, { useState, useRef, useEffect } from 'react';
import { useNetwork, AppPage } from '../../context/NetworkContext';
import { useTheme } from '../../context/ThemeContext';
import { useI18n } from '../../context/I18nContext';
import {
  RotateCw,
  Radar,
  Wifi,
  ChevronRight,
  ChevronLeft,
  Sun,
  Moon,
  Menu,
  Languages,
  Check
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
  const { t, language, setLanguage, isRTL } = useI18n();
  const [langMenuOpen, setLangMenuOpen] = useState(false);
  const langMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (langMenuRef.current && !langMenuRef.current.contains(e.target as Node)) {
        setLangMenuOpen(false);
      }
    };
    if (langMenuOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [langMenuOpen]);

  const availableLanguages: { code: 'en' | 'ar'; label: string; nativeLabel: string; dir: 'ltr' | 'rtl' }[] = [
    { code: 'en', label: 'English', nativeLabel: 'English', dir: 'ltr' },
    { code: 'ar', label: 'Arabic', nativeLabel: 'العربية', dir: 'rtl' },
  ];

  const getPageTitle = () => {
    switch (activePage) {
      case 'dashboard':
        return t('nav.dashboard');
      case 'devices':
        return t('devices.title');
      case 'device-details':
        return t('deviceDetails.netIdCard');
      case 'network':
        return t('network.title');
      case 'scanner':
        return t('scanner.title');
      case 'events':
        return t('events.title');
      case 'settings':
        return t('settings.title');
      default:
        return 'NETWATCH';
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
            {isRTL ? <ChevronLeft className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
            <button
              onClick={() => navigateTo('devices')}
              className="hover:underline hover:text-neutral-700 dark:hover:text-neutral-200"
            >
              {t('nav.devices')}
            </button>
            {isRTL ? <ChevronLeft className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
            <span className="font-mono text-neutral-700 dark:text-neutral-300 font-medium">
              {selectedDevice.name}
            </span>
          </div>
        )}
      </div>

      {/* Right: Current network indicator & primary action group with reduced spacing */}
      <div className="flex items-center gap-1.5">
        {/* Network & Subnet indicator */}
        <div className="hidden md:flex items-center gap-1.5 text-xs text-neutral-600 dark:text-neutral-400 px-2.5 py-1.5 rounded bg-neutral-100 dark:bg-neutral-800/80 border border-neutral-200 dark:border-neutral-700 font-mono me-1">
          <Wifi className="w-3.5 h-3.5 text-neutral-500" />
          <span className="font-sans font-medium text-neutral-800 dark:text-neutral-200">
            {networkInfo?.ssid || networkInfo?.interfaceName || t('common.disconnected')}
          </span>
          <span className="text-neutral-400 dark:text-neutral-600">·</span>
          <span dir="ltr">{networkInfo?.subnet || '192.168.1.0/24'}</span>
          <span className="text-neutral-400 dark:text-neutral-600">·</span>
          <span className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400 font-sans font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            {t('common.connected')}
          </span>
        </div>

        {/* Refresh button - square */}
        <button
          onClick={() => refresh()}
          disabled={isScanning}
          className="p-1.5 text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded border border-neutral-200 dark:border-neutral-700 transition-colors disabled:opacity-50 cursor-pointer"
          title={t('common.rescan')}
          aria-label={t('common.rescan')}
        >
          <RotateCw className="w-4 h-4" />
        </button>

        {/* Theme Toggle Button - square */}
        <button
          onClick={toggleTheme}
          className="p-1.5 text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded border border-neutral-200 dark:border-neutral-700 transition-colors cursor-pointer"
          title={t('topbar.themeToggle')}
          aria-label={t('topbar.themeToggle')}
        >
          {theme === 'dark' ? (
            <Sun className="w-4 h-4 text-amber-400" />
          ) : (
            <Moon className="w-4 h-4 text-neutral-600" />
          )}
        </button>

        {/* Language Switcher - compact square button with dropdown */}
        <div className="relative" ref={langMenuRef}>
          <button
            onClick={() => setLangMenuOpen(prev => !prev)}
            className={`p-1.5 rounded border transition-colors cursor-pointer flex items-center justify-center ${
              langMenuOpen
                ? 'bg-neutral-200 dark:bg-neutral-700 text-neutral-950 dark:text-white border-neutral-300 dark:border-neutral-600 shadow-2xs'
                : 'text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-100 dark:hover:bg-neutral-800 border-neutral-200 dark:border-neutral-700'
            }`}
            title={t('topbar.languageToggle')}
            aria-label={t('topbar.languageToggle')}
            aria-expanded={langMenuOpen}
          >
            <Languages className="w-4 h-4" />
          </button>

          {/* Dropdown Menu */}
          {langMenuOpen && (
            <div
              className={`absolute top-full mt-1.5 w-36 bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg shadow-lg py-1 z-50 ${
                isRTL ? 'left-0' : 'right-0'
              }`}
            >
              <div className="px-3 py-1 text-[10px] font-semibold text-neutral-400 dark:text-neutral-500 uppercase tracking-wider border-b border-neutral-100 dark:border-neutral-800/80 mb-0.5">
                {t('topbar.languageToggle')}
              </div>
              {availableLanguages.map(lang => (
                <button
                  key={lang.code}
                  onClick={() => {
                    setLanguage(lang.code);
                    setLangMenuOpen(false);
                  }}
                  className={`w-full px-3 py-1.5 text-xs text-start flex items-center justify-between transition-colors cursor-pointer ${
                    language === lang.code
                      ? 'bg-neutral-100 dark:bg-neutral-800/80 text-neutral-900 dark:text-neutral-100 font-semibold'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-50 dark:hover:bg-neutral-800/40 hover:text-neutral-900 dark:hover:text-neutral-100'
                  }`}
                  dir={lang.dir}
                >
                  <span>{lang.nativeLabel}</span>
                  {language === lang.code && (
                    <Check className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400 shrink-0" />
                  )}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Global Scan Action */}
        {activePage !== 'scanner' && (
          <button
            onClick={() => {
              navigateTo('scanner');
              startScan('quick');
            }}
            disabled={isScanning}
            className="px-3 py-1.5 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-950 hover:bg-neutral-800 dark:hover:bg-white text-xs font-semibold rounded transition-colors shadow-xs flex items-center gap-1.5 disabled:opacity-60 disabled:cursor-not-allowed cursor-pointer"
          >
            {isScanning ? (
              <>
                <RotateCw className="w-3.5 h-3.5 animate-spin" />
                <span>{t('common.scanning')} {scanProgress.progress}%</span>
              </>
            ) : (
              <>
                <Radar className="w-3.5 h-3.5 text-emerald-500 dark:text-emerald-600" />
                <span>{t('common.scan')}</span>
              </>
            )}
          </button>
        )}
      </div>
    </div>
  );
};
