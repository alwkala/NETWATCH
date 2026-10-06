import React, { useState, useMemo, useDeferredValue } from 'react';
import { useNetwork } from '../context/NetworkContext';
import { EventType } from '../types/events';
import {
  Clock,
  Search,
  Filter,
  ArrowRight,
  HardDrive,
  Trash2
} from 'lucide-react';

export const Events: React.FC = () => {
  const { events, selectDevice, devices } = useNetwork();
  const [filterType, setFilterType] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const deferredSearch = useDeferredValue(searchQuery);

  const filterTabs = [
    { id: 'all', label: 'All Events' },
    { id: 'new_device', label: 'New Device' },
    { id: 'online', label: 'Online' },
    { id: 'offline', label: 'Offline' },
    { id: 'network_change', label: 'Network Change' },
    { id: 'scan', label: 'Scan' }
  ];

  const filteredEvents = useMemo(() => {
    let list = [...events];

    if (filterType !== 'all') {
      list = list.filter(e => e.type === filterType);
    }

    if (deferredSearch.trim()) {
      const q = deferredSearch.toLowerCase().trim();
      list = list.filter(
        e =>
          e.title.toLowerCase().includes(q) ||
          e.deviceName?.toLowerCase().includes(q) ||
          e.ip?.toLowerCase().includes(q) ||
          e.details?.toLowerCase().includes(q)
      );
    }

    return list;
  }, [events, filterType, deferredSearch]);

  return (
    <div className="p-6 space-y-4 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-neutral-200 dark:border-neutral-800">
        <div>
          <h2 className="text-lg font-bold text-neutral-900 dark:text-neutral-100">
            Network Activity Log
          </h2>
          <p className="text-xs text-neutral-500 dark:text-neutral-400">
            Audit log of IP changes, device appearances, heartbeats, and scanner events.
          </p>
        </div>

        <div className="text-xs font-mono text-neutral-400">
          {filteredEvents.length} events logged
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="relative flex-1 min-w-50 max-w-sm">
          <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-neutral-400" />
          <input
            type="text"
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            placeholder="Search events by device, IP, or title..."
            className="w-full bg-white dark:bg-neutral-800 text-xs text-neutral-900 dark:text-neutral-100 pl-8 pr-3 py-1.5 rounded border border-neutral-300 dark:border-neutral-700 focus:outline-hidden"
          />
        </div>

        {/* Filter buttons */}
        <div className="flex flex-wrap items-center gap-1 bg-neutral-100 dark:bg-neutral-800 p-1 rounded border border-neutral-200 dark:border-neutral-700">
          {filterTabs.map(tab => (
            <button
              key={tab.id}
              onClick={() => setFilterType(tab.id)}
              className={`px-2.5 py-1 text-[11px] font-medium rounded transition-colors ${
                filterType === tab.id
                  ? 'bg-white dark:bg-neutral-700 text-neutral-900 dark:text-neutral-100 shadow-xs'
                  : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-200'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Events Table */}
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md overflow-hidden shadow-xs transition-colors">
        {filteredEvents.length === 0 ? (
          /* Empty state (Section 23) */
          <div className="p-12 text-center text-neutral-400">
            <Clock className="w-8 h-8 mx-auto text-neutral-300 dark:text-neutral-600 mb-2" />
            <p className="text-xs font-semibold text-neutral-800 dark:text-neutral-200">
              No network events yet
            </p>
            <p className="text-[11px] text-neutral-500 mt-0.5">
              Network state transitions and discoveries will appear here automatically.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="bg-neutral-50 dark:bg-neutral-800/70 border-b border-neutral-200 dark:border-neutral-800 text-neutral-500 uppercase tracking-wider text-[10px]">
                  <th className="py-2.5 px-4 w-28">Time</th>
                  <th className="py-2.5 px-4 w-36">Event</th>
                  <th className="py-2.5 px-4">Device / Target</th>
                  <th className="py-2.5 px-4 hidden md:table-cell">Details</th>
                  <th className="py-2.5 px-3 w-10" />
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800">
                {filteredEvents.map(evt => {
                  const hasLinkedDevice = evt.deviceId || devices.some(d => d.ip === evt.ip);
                  const linkedDeviceId = evt.deviceId || devices.find(d => d.ip === evt.ip)?.id;

                  return (
                    <tr
                      key={evt.id}
                      onClick={() => {
                        if (linkedDeviceId) selectDevice(linkedDeviceId);
                      }}
                      className={`hover:bg-neutral-50 dark:hover:bg-neutral-800/60 transition-colors ${
                        linkedDeviceId ? 'cursor-pointer group' : ''
                      }`}
                    >
                      <td className="py-3 px-4 font-mono text-neutral-500 dark:text-neutral-400 tabular-nums">
                        {evt.timestamp}
                      </td>

                      <td className="py-3 px-4">
                        <span className="inline-flex items-center gap-1.5 font-medium">
                          <span
                            className={`w-2 h-2 rounded-full shrink-0 ${
                              evt.type === 'new_device'
                                ? 'bg-amber-500'
                                : evt.type === 'online'
                                ? 'bg-emerald-500'
                                : evt.type === 'offline'
                                ? 'bg-neutral-400'
                                : 'bg-sky-500'
                            }`}
                          />
                          <span className="capitalize">{evt.title}</span>
                        </span>
                      </td>

                      <td className="py-3 px-4">
                        {evt.deviceName ? (
                          <div>
                            <div className="font-semibold text-neutral-900 dark:text-neutral-100 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                              {evt.deviceName}
                            </div>
                            {evt.ip && (
                              <div className="font-mono text-[11px] text-neutral-500">
                                {evt.ip}
                              </div>
                            )}
                          </div>
                        ) : (
                          <span className="text-neutral-400">—</span>
                        )}
                      </td>

                      <td className="py-3 px-4 text-neutral-600 dark:text-neutral-400 hidden md:table-cell text-[11px]">
                        {evt.details}
                      </td>

                      <td className="py-3 px-3 text-right">
                        {linkedDeviceId && (
                          <ArrowRight className="w-3.5 h-3.5 text-neutral-400 group-hover:text-neutral-800 dark:group-hover:text-white transition-colors" />
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
