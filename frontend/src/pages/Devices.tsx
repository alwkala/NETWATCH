import React from 'react';
import { useNetwork } from '../context/NetworkContext';
import { useI18n } from '../context/I18nContext';
import { DeviceTable } from '../components/devices/DeviceTable';
import { TableLoadingSkeleton } from '../components/ui/LoadingSkeleton';
import { ErrorState } from '../components/ui/ErrorState';
import {
  RotateCw,
  Radar,
  Plus,
  Radio
} from 'lucide-react';

export const Devices: React.FC = () => {
  const {
    devices,
    networkInfo,
    selectDevice,
    startScan,
    addNewSimulatedDevice,
    service,
    isScanning,
    isLoading,
    error,
    clearError
  } = useNetwork();
  const { t } = useI18n();

  if (isLoading) {
    return <TableLoadingSkeleton />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={clearError} />;
  }

  const onlineCount = devices.filter(d => d.status === 'online').length;
  const offlineCount = devices.filter(d => d.status === 'offline').length;
  const newCount = devices.filter(d => d.isNew).length;

  if (devices.length === 0) {
    return (
      <div className="p-12 text-center max-w-md mx-auto min-h-[50vh] flex flex-col items-center justify-center">
        <Radio className="w-10 h-10 text-neutral-400 mb-3" />
        <h3 className="text-base font-bold text-neutral-900 dark:text-neutral-100 mb-1">
          {t('devices.emptyTitle')}
        </h3>
        <p className="text-xs text-neutral-500 dark:text-neutral-400 mb-4">
          {t('devices.scanPrompt')}
        </p>
        <button
          onClick={() => startScan('quick')}
          disabled={isScanning}
          className="px-4 py-2 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-900 text-xs font-semibold rounded flex items-center gap-2 hover:bg-neutral-800 dark:hover:bg-white transition-colors"
        >
          <Radar className="w-3.5 h-3.5 text-emerald-500" />
          <span>{t('common.scan')}</span>
        </button>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4 max-w-7xl mx-auto">
      {/* Header bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-neutral-200 dark:border-neutral-800">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-bold text-neutral-900 dark:text-neutral-100">
              {t('devices.title')}
            </h2>
            <span className="text-xs font-mono text-neutral-500 px-2 py-0.5 bg-neutral-100 dark:bg-neutral-800 rounded">
              {devices.length} {t('nav.devices')}
            </span>
          </div>
          <p className="text-xs text-neutral-500 dark:text-neutral-400">
            {onlineCount} {t('common.online')} · {offlineCount} {t('common.offline')} · {newCount} {t('devices.newBadge')} · <span dir="ltr">{networkInfo?.subnet || '192.168.1.0/24'}</span>
          </p>
        </div>

        {/* Action controls (simulation only in prototype) */}
        {service.isSimulated && (
          <div className="flex items-center gap-2">
            <button
              onClick={() => addNewSimulatedDevice()}
              className="px-3 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-700 dark:text-neutral-300 rounded text-xs font-medium border border-neutral-200 dark:border-neutral-700 flex items-center gap-1.5 transition-colors"
              title="Simulate detecting a new hardware device"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Simulate New Device</span>
            </button>
          </div>
        )}
      </div>

      {/* Main Device Table with Full Search & Sorting */}
      <DeviceTable
        devices={devices}
        onSelectDevice={dev => selectDevice(dev.id)}
        showControls={true}
        compact={false}
      />
    </div>
  );
};
