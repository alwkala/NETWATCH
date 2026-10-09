import React, { useState, useMemo, useDeferredValue } from 'react';
import { Device, DeviceType } from '../../types/device';
import { DeviceStatusBadge } from './DeviceStatusBadge';
import { DeviceTypeIcon } from './DeviceTypeIcon';
import { useI18n } from '../../context/I18nContext';
import {
  Search,
  ArrowUpDown,
  Filter,
  X,
  ChevronRight,
  ChevronLeft,
  HardDrive
} from 'lucide-react';

interface DeviceTableProps {
  devices: Device[];
  onSelectDevice: (device: Device) => void;
  compact?: boolean;
  limit?: number;
  initialFilter?: 'all' | 'online' | 'offline' | 'new';
  showControls?: boolean;
}

type SortField = 'name' | 'ip' | 'vendor' | 'type' | 'latency' | 'lastSeen' | 'status';

export const DeviceTable: React.FC<DeviceTableProps> = ({
  devices,
  onSelectDevice,
  compact = false,
  limit,
  initialFilter = 'all',
  showControls = true
}) => {
  const { t, isRTL } = useI18n();
  const [searchQuery, setSearchQuery] = useState('');
  const deferredSearch = useDeferredValue(searchQuery);
  const [statusFilter, setStatusFilter] = useState<'all' | 'online' | 'offline' | 'new'>(initialFilter);
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [trustFilter, setTrustFilter] = useState<string>('all');
  const [sortField, setSortField] = useState<SortField>('ip');
  const [sortAsc, setSortAsc] = useState<boolean>(true);

  // Extract unique device types
  const availableTypes = useMemo(() => {
    const set = new Set<string>();
    devices.forEach(d => set.add(d.type));
    return Array.from(set).sort();
  }, [devices]);

  // Filtered & Sorted devices
  const processedDevices = useMemo(() => {
    let result = [...devices];

    // Status filter
    if (statusFilter === 'online') {
      result = result.filter(d => d.status === 'online');
    } else if (statusFilter === 'offline') {
      result = result.filter(d => d.status === 'offline');
    } else if (statusFilter === 'new') {
      result = result.filter(d => d.isNew);
    }

    // Type filter
    if (typeFilter !== 'all') {
      result = result.filter(d => d.type === typeFilter);
    }

    // Trust filter
    if (trustFilter !== 'all') {
      result = result.filter(d => (d.trustStatus || 'unknown') === trustFilter);
    }

    // Search query (name, ip, mac, vendor, hostname)
    if (deferredSearch.trim()) {
      const q = deferredSearch.toLowerCase().trim();
      result = result.filter(
        d =>
          d.name.toLowerCase().includes(q) ||
          d.hostname.toLowerCase().includes(q) ||
          d.ip.toLowerCase().includes(q) ||
          d.mac.toLowerCase().includes(q) ||
          d.vendor.toLowerCase().includes(q)
      );
    }

    // Sort
    result.sort((a, b) => {
      let comp = 0;
      switch (sortField) {
        case 'name':
          comp = a.name.localeCompare(b.name);
          break;
        case 'ip': {
          // IP sort numeric by last octet or all octets
          const ipA = a.ip.split('.').map(Number);
          const ipB = b.ip.split('.').map(Number);
          for (let i = 0; i < 4; i++) {
            if (ipA[i] !== ipB[i]) {
              comp = (ipA[i] || 0) - (ipB[i] || 0);
              break;
            }
          }
          break;
        }
        case 'vendor':
          comp = a.vendor.localeCompare(b.vendor);
          break;
        case 'type':
          comp = a.type.localeCompare(b.type);
          break;
        case 'latency':
          comp = (a.latencyMs || 9999) - (b.latencyMs || 9999);
          break;
        case 'lastSeen':
          comp = a.lastSeenAt && b.lastSeenAt
            ? a.lastSeenAt.localeCompare(b.lastSeenAt)
            : a.lastSeen.localeCompare(b.lastSeen);
          break;
        case 'status':
          comp = a.status.localeCompare(b.status);
          break;
      }
      return sortAsc ? comp : -comp;
    });

    if (limit && limit > 0) {
      result = result.slice(0, limit);
    }

    return result;
  }, [devices, statusFilter, typeFilter, deferredSearch, sortField, sortAsc, limit]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortAsc(!sortAsc);
    } else {
      setSortField(field);
      setSortAsc(true);
    }
  };

  return (
    <div className="w-full">
      {/* Controls & Filter Bar */}
      {showControls && (
        <div className="flex flex-wrap items-center justify-between gap-3 mb-3">
          {/* Search box */}
          <div className="relative flex-1 min-w-50 max-w-sm">
            <Search className={`w-3.5 h-3.5 absolute ${isRTL ? 'right-2.5' : 'left-2.5'} top-1/2 -translate-y-1/2 text-neutral-400`} />
            <input
              type="text"
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              placeholder={t('devices.search')}
              className={`w-full bg-white dark:bg-neutral-800 text-xs text-neutral-900 dark:text-neutral-100 ${isRTL ? 'pr-8 pl-7' : 'pl-8 pr-7'} py-1.5 rounded border border-neutral-300 dark:border-neutral-700 focus:outline-hidden focus:border-neutral-500 placeholder:text-neutral-400 transition-colors`}
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className={`absolute ${isRTL ? 'left-2' : 'right-2'} top-1/2 -translate-y-1/2 text-neutral-400 hover:text-neutral-600`}
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>

          {/* Status Tabs (Interactive filter buttons) */}
          <div className="flex items-center gap-1 bg-neutral-100 dark:bg-neutral-800 p-1 rounded border border-neutral-200 dark:border-neutral-700">
            {(['all', 'online', 'offline', 'new'] as const).map(tab => {
              const label = tab === 'all'
                ? t('devices.filterAll')
                : tab === 'online'
                ? t('devices.filterOnline')
                : tab === 'offline'
                ? t('devices.filterOffline')
                : t('devices.newBadge');
              return (
                <button
                  key={tab}
                  onClick={() => setStatusFilter(tab)}
                  className={`px-2.5 py-1 text-xs font-medium rounded transition-colors text-[10px] ${
                    statusFilter === tab
                      ? 'bg-white dark:bg-neutral-700 text-neutral-900 dark:text-neutral-100 shadow-xs font-semibold'
                      : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-200'
                  }`}
                >
                  {label}
                </button>
              );
            })}
          </div>

          {/* Type Filter dropdown if not compact */}
          {!compact && (
            <div className="flex items-center gap-2">
              <span className="text-xs text-neutral-400 flex items-center gap-1">
                <Filter className="w-3 h-3" />
                {t('devices.filterType')}:
              </span>
              <select
                value={typeFilter}
                onChange={e => setTypeFilter(e.target.value)}
                className="bg-white dark:bg-neutral-800 text-xs text-neutral-900 dark:text-neutral-200 border border-neutral-300 dark:border-neutral-700 rounded px-2 py-1 focus:outline-hidden"
              >
                <option value="all">{t('devices.filterAll')}</option>
                {availableTypes.map(t => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
              </select>

              <select
                value={trustFilter}
                onChange={e => setTrustFilter(e.target.value)}
                className="bg-white dark:bg-neutral-800 text-xs text-neutral-900 dark:text-neutral-200 border border-neutral-300 dark:border-neutral-700 rounded px-2 py-1 focus:outline-hidden"
              >
                <option value="all">{t('devices.filterTrust')}</option>
                <option value="known">{t('devices.trustKnown')}</option>
                <option value="guest">{t('devices.trustGuest')}</option>
                <option value="unknown">{t('devices.trustUnknown')}</option>
              </select>
            </div>
          )}
        </div>
      )}

      {/* Device Table */}
      <div className="border border-neutral-200 dark:border-neutral-800 rounded-md overflow-hidden bg-white dark:bg-neutral-900 shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-left rtl:text-right border-collapse text-xs">
            <thead>
              <tr className="bg-neutral-50 dark:bg-neutral-800/70 border-b border-neutral-200 dark:border-neutral-800 text-neutral-500 dark:text-neutral-400 font-semibold uppercase tracking-wider text-[10px]">
                <th
                  onClick={() => handleSort('status')}
                  className="py-2.5 px-3 cursor-pointer hover:text-neutral-900 dark:hover:text-neutral-200 w-24"
                >
                  <div className="flex items-center gap-1">
                    <span>{t('devices.colStatus')}</span>
                    <ArrowUpDown className="w-3 h-3 opacity-60" />
                  </div>
                </th>
                <th
                  onClick={() => handleSort('name')}
                  className="py-2.5 px-3 cursor-pointer hover:text-neutral-900 dark:hover:text-neutral-200"
                >
                  <div className="flex items-center gap-1">
                    <span>{t('devices.colName')}</span>
                    <ArrowUpDown className="w-3 h-3 opacity-60" />
                  </div>
                </th>
                <th
                  onClick={() => handleSort('ip')}
                  className="py-2.5 px-3 cursor-pointer hover:text-neutral-900 dark:hover:text-neutral-200 w-32"
                >
                  <div className="flex items-center gap-1">
                    <span>{t('devices.colIP')}</span>
                    <ArrowUpDown className="w-3 h-3 opacity-60" />
                  </div>
                </th>
                {!compact && (
                  <th className="py-2.5 px-3 text-neutral-400 w-36 hidden lg:table-cell">
                    {t('devices.colMAC')}
                  </th>
                )}
                <th
                  onClick={() => handleSort('vendor')}
                  className="py-2.5 px-3 cursor-pointer hover:text-neutral-900 dark:hover:text-neutral-200 hidden md:table-cell"
                >
                  <div className="flex items-center gap-1">
                    <span>{t('devices.colVendor')}</span>
                    <ArrowUpDown className="w-3 h-3 opacity-60" />
                  </div>
                </th>
                <th
                  onClick={() => handleSort('type')}
                  className="py-2.5 px-3 cursor-pointer hover:text-neutral-900 dark:hover:text-neutral-200"
                >
                  <div className="flex items-center gap-1">
                    <span>{t('devices.colType')}</span>
                    <ArrowUpDown className="w-3 h-3 opacity-60" />
                  </div>
                </th>
                <th
                  onClick={() => handleSort('latency')}
                  className="py-2.5 px-3 cursor-pointer hover:text-neutral-900 dark:hover:text-neutral-200 text-right rtl:text-left w-20 hidden sm:table-cell"
                >
                  <div className="flex items-center justify-end rtl:justify-start gap-1">
                    <span>{t('devices.colLatency')}</span>
                    <ArrowUpDown className="w-3 h-3 opacity-60" />
                  </div>
                </th>
                <th
                  onClick={() => handleSort('lastSeen')}
                  className="py-2.5 px-3 cursor-pointer hover:text-neutral-900 dark:hover:text-neutral-200 text-right rtl:text-left w-24 hidden md:table-cell"
                >
                  <div className="flex items-center justify-end rtl:justify-start gap-1">
                    <span>{t('devices.colLastSeen')}</span>
                    <ArrowUpDown className="w-3 h-3 opacity-60" />
                  </div>
                </th>
                <th className="py-2.5 px-2 w-8" />
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800">
              {processedDevices.length === 0 ? (
                <tr>
                  <td colSpan={compact ? 5 : 9} className="py-12 text-center text-neutral-400">
                    <div className="flex flex-col items-center justify-center">
                      <HardDrive className="w-6 h-6 text-neutral-300 dark:text-neutral-600 mb-2" />
                      <p className="text-xs font-medium text-neutral-700 dark:text-neutral-300">
                        {t('devices.emptyTitle')}
                      </p>
                      <p className="text-[11px] text-neutral-400 mt-0.5">
                        {t('devices.emptyDesc')}
                      </p>
                    </div>
                  </td>
                </tr>
              ) : (
                processedDevices.map(device => (
                  <tr
                    key={device.id}
                    onClick={() => onSelectDevice(device)}
                    className="hover:bg-neutral-50 dark:hover:bg-neutral-800/60 cursor-pointer transition-colors group"
                  >
                    {/* Status */}
                    <td className="py-2.5 px-3 whitespace-nowrap">
                      <DeviceStatusBadge
                        status={device.status}
                        isNew={device.isNew}
                      />
                    </td>

                    {/* Device & Hostname */}
                    <td className="py-2.5 px-3">
                      <div className="flex items-center gap-1.5 flex-wrap">
                        <span className="font-semibold text-neutral-900 dark:text-neutral-100 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                          {device.customAlias || device.name}
                        </span>
                        {device.trustStatus === 'known' && (
                          <span className="text-[10px] px-1.5 py-0.2 bg-emerald-50 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800/80 rounded font-medium">
                            {t('devices.trustKnown')}
                          </span>
                        )}
                        {device.trustStatus === 'guest' && (
                          <span className="text-[10px] px-1.5 py-0.2 bg-blue-50 dark:bg-blue-950/60 text-blue-700 dark:text-blue-400 border border-blue-200 dark:border-blue-800/80 rounded font-medium">
                            {t('devices.trustGuest')}
                          </span>
                        )}
                        {device.isRandomizedMac && (
                          <span className="text-[10px] px-1.5 py-0.2 bg-purple-50 dark:bg-purple-950/60 text-purple-700 dark:text-purple-400 border border-purple-200 dark:border-purple-800/80 rounded font-mono" title={t('devices.privateMACTooltip')}>
                            {t('devices.privateMAC')}
                          </span>
                        )}
                      </div>
                      <div className="text-[11px] text-neutral-500 dark:text-neutral-400 font-mono truncate max-w-45" dir="ltr">
                        {device.hostname || '—'}
                      </div>
                    </td>

                    {/* IP Address */}
                    <td className="py-2.5 px-3 font-mono font-medium text-neutral-800 dark:text-neutral-200 tabular-nums whitespace-nowrap" dir="ltr">
                      {device.ip}
                    </td>

                    {/* MAC Address */}
                    {!compact && (
                      <td className="py-2.5 px-3 font-mono text-[11px] text-neutral-500 dark:text-neutral-400 whitespace-nowrap hidden lg:table-cell" dir="ltr">
                        {device.mac}
                      </td>
                    )}

                    {/* Vendor */}
                    <td className="py-2.5 px-3 text-neutral-600 dark:text-neutral-400 truncate max-w-35 hidden md:table-cell">
                      {device.vendor}
                    </td>

                    {/* Type with icon */}
                    <td className="py-2.5 px-3 whitespace-nowrap">
                      <div className="flex items-center gap-1.5 text-neutral-600 dark:text-neutral-300">
                        <DeviceTypeIcon
                          type={device.type}
                          className="w-3.5 h-3.5 text-neutral-500 shrink-0"
                        />
                        <span>{device.type}</span>
                      </div>
                    </td>

                    {/* Latency */}
                    <td className="py-2.5 px-3 font-mono text-right rtl:text-left text-neutral-700 dark:text-neutral-300 tabular-nums whitespace-nowrap hidden sm:table-cell" dir="ltr">
                      {device.latencyMs !== undefined ? `${device.latencyMs} ms` : '—'}
                    </td>

                    {/* Last Seen */}
                    <td className="py-2.5 px-3 text-right rtl:text-left text-neutral-500 dark:text-neutral-400 whitespace-nowrap hidden md:table-cell" dir="ltr">
                      {device.lastSeen}
                    </td>

                    {/* Action Arrow */}
                    <td className="py-2.5 px-2 text-right rtl:text-left">
                      {isRTL ? (
                        <ChevronLeft className="w-3.5 h-3.5 text-neutral-300 dark:text-neutral-600 group-hover:text-neutral-700 dark:group-hover:text-neutral-300 transition-colors" />
                      ) : (
                        <ChevronRight className="w-3.5 h-3.5 text-neutral-300 dark:text-neutral-600 group-hover:text-neutral-700 dark:group-hover:text-neutral-300 transition-colors" />
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
