import React, { useState } from 'react';
import { useNetwork } from '../context/NetworkContext';
import { DeviceStatusBadge } from '../components/devices/DeviceStatusBadge';
import { DeviceTypeIcon } from '../components/devices/DeviceTypeIcon';
import {
  ArrowLeft,
  Activity,
  Zap,
  Terminal,
  Edit2,
  Check,
  X,
  RotateCw,
  Clock,
  Shield,
  Layers,
  Power
} from 'lucide-react';

export const DeviceDetails: React.FC = () => {
  const {
    selectedDevice,
    navigateTo,
    pingDevice,
    wakeOnLan,
    scanDevicePorts,
    updateDevice,
    toggleDeviceStatus,
    service
  } = useNetwork();

  const [isPinging, setIsPinging] = useState(false);
  const [pingResult, setPingResult] = useState<string | null>(null);
  const [isWaking, setIsWaking] = useState(false);
  const [wolResult, setWolResult] = useState<string | null>(null);
  const [isScanningPorts, setIsScanningPorts] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [customAlias, setCustomAlias] = useState('');
  const [notes, setNotes] = useState('');

  if (!selectedDevice) {
    return (
      <div className="p-8 text-center">
        <p className="text-xs text-neutral-400 mb-4">No device selected.</p>
        <button
          onClick={() => navigateTo('devices')}
          className="text-xs font-semibold underline text-neutral-700 dark:text-neutral-300"
        >
          Return to Devices
        </button>
      </div>
    );
  }

  const handleStartEdit = () => {
    setCustomAlias(selectedDevice.customAlias || selectedDevice.name);
    setNotes(selectedDevice.notes || '');
    setIsEditing(true);
  };

  const handleSaveEdit = async () => {
    await updateDevice(selectedDevice.id, {
      customAlias: customAlias.trim() || undefined,
      notes: notes.trim() || undefined
    });
    setIsEditing(false);
  };

  const handlePing = async () => {
    setIsPinging(true);
    setPingResult(null);
    try {
      const res = await pingDevice(selectedDevice.ip);
      if (res.success) {
        setPingResult(`ICMP reply from ${selectedDevice.ip}: time=${res.latencyMs}ms TTL=64`);
      } else {
        setPingResult(`Request timed out for ${selectedDevice.ip}. Host unreachable.`);
      }
    } catch {
      setPingResult('Ping failed: network unreachable.');
    } finally {
      setIsPinging(false);
    }
  };

  const handleWol = async () => {
    setIsWaking(true);
    setWolResult(null);
    try {
      const res = await wakeOnLan(selectedDevice.mac);
      setWolResult(res.message);
    } catch {
      setWolResult('Failed to broadcast Wake-on-LAN packet.');
    } finally {
      setIsWaking(false);
    }
  };

  const handleScanServices = async () => {
    setIsScanningPorts(true);
    try {
      await scanDevicePorts(selectedDevice.id);
    } finally {
      setIsScanningPorts(false);
    }
  };

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      {/* Back button */}
      <div>
        <button
          onClick={() => navigateTo('devices')}
          className="inline-flex items-center gap-1.5 text-xs text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 transition-colors"
        >
          <ArrowLeft className="w-3.5 h-3.5" />
          <span>Back to Devices</span>
        </button>
      </div>

      {/* Device Header */}
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 shadow-xs transition-colors">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="flex items-start gap-3.5">
            <div className="w-11 h-11 rounded-md bg-neutral-100 dark:bg-neutral-800 flex items-center justify-center text-neutral-700 dark:text-neutral-300 border border-neutral-200 dark:border-neutral-700">
              <DeviceTypeIcon type={selectedDevice.type} className="w-5 h-5" />
            </div>

            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-xl font-bold text-neutral-900 dark:text-neutral-100">
                  {selectedDevice.customAlias || selectedDevice.name}
                </h2>
                <DeviceStatusBadge
                  status={selectedDevice.status}
                  isNew={selectedDevice.isNew}
                />
              </div>

              <div className="flex flex-wrap items-center gap-2 text-xs font-mono text-neutral-500 dark:text-neutral-400 mt-1">
                <span className="text-neutral-900 dark:text-neutral-200 font-semibold">{selectedDevice.ip}</span>
                <span>·</span>
                <span>{selectedDevice.mac}</span>
                <span>·</span>
                <span className="font-sans text-neutral-600 dark:text-neutral-400">{selectedDevice.vendor}</span>
              </div>
            </div>
          </div>

          {/* Action buttons (Ping, WOL, Scan, Edit) */}
          <div className="flex flex-wrap items-center gap-2">
            <button
              onClick={handlePing}
              disabled={isPinging}
              className="px-3 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-800 dark:text-neutral-200 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 flex items-center gap-1.5 transition-colors disabled:opacity-50"
              title="Send ICMP echo ping"
            >
              {isPinging ? (
                <RotateCw className="w-3.5 h-3.5 animate-spin" />
              ) : (
                <Activity className="w-3.5 h-3.5 text-emerald-500" />
              )}
              <span>Ping</span>
            </button>

            <button
              onClick={handleWol}
              disabled={isWaking}
              className="px-3 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-800 dark:text-neutral-200 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 flex items-center gap-1.5 transition-colors disabled:opacity-50"
              title="Broadcast Wake-on-LAN magic packet"
            >
              <Zap className="w-3.5 h-3.5 text-amber-500" />
              <span>Wake on LAN</span>
            </button>

            <button
              onClick={handleScanServices}
              disabled={isScanningPorts || selectedDevice.status === 'offline'}
              className="px-3 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-800 dark:text-neutral-200 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 flex items-center gap-1.5 transition-colors disabled:opacity-50"
              title="Scan TCP/UDP service ports"
            >
              {isScanningPorts ? (
                <RotateCw className="w-3.5 h-3.5 animate-spin" />
              ) : (
                <Terminal className="w-3.5 h-3.5" />
              )}
              <span>Scan Ports</span>
            </button>

            <button
              onClick={handleStartEdit}
              className="px-3 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-800 dark:text-neutral-200 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 flex items-center gap-1.5 transition-colors"
            >
              <Edit2 className="w-3.5 h-3.5" />
              <span>Edit</span>
            </button>

            {selectedDevice.isNew && (
              <button
                onClick={() => updateDevice(selectedDevice.id, { isNew: false })}
                className="px-3 py-1.5 bg-amber-50 dark:bg-amber-950/40 hover:bg-amber-100 dark:hover:bg-amber-900/40 text-amber-800 dark:text-amber-300 rounded text-xs font-medium border border-amber-300 dark:border-amber-800 transition-colors"
                title="Mark this device as known"
              >
                Acknowledge
              </button>
            )}

            {/* Quick state toggle for testing (prototype only) */}
            {service.isSimulated && (<button
              onClick={() => toggleDeviceStatus(selectedDevice.id)}
              className="p-1.5 text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded border border-neutral-300 dark:border-neutral-700"
              title={`Simulate device going ${selectedDevice.status === 'online' ? 'Offline' : 'Online'}`}
            >
              <Power className="w-3.5 h-3.5" />
            </button>)}
          </div>
        </div>

        {/* Live Ping or WOL feedback message */}
        {pingResult && (
          <div className="mt-3 p-2.5 rounded bg-neutral-900 text-emerald-400 font-mono text-xs flex items-center justify-between">
            <span>{pingResult}</span>
            <button onClick={() => setPingResult(null)} className="text-neutral-400 hover:text-white">
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        )}
        {wolResult && (
          <div className="mt-3 p-2.5 rounded bg-neutral-900 text-amber-300 font-mono text-xs flex items-center justify-between">
            <span>{wolResult}</span>
            <button onClick={() => setWolResult(null)} className="text-neutral-400 hover:text-white">
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        )}

        {/* Edit Modal / Drawer */}
        {isEditing && (
          <div className="mt-4 pt-4 border-t border-neutral-200 dark:border-neutral-800 space-y-3">
            <div className="text-xs font-semibold text-neutral-700 dark:text-neutral-300">
              Edit Device Details
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label className="text-[11px] text-neutral-500 block mb-1">
                  Custom Friendly Name
                </label>
                <input
                  type="text"
                  value={customAlias}
                  onChange={e => setCustomAlias(e.target.value)}
                  className="w-full text-xs p-1.5 rounded bg-neutral-50 dark:bg-neutral-900 border border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100"
                />
              </div>
              <div>
                <label className="text-[11px] text-neutral-500 block mb-1">
                  Local Notes
                </label>
                <input
                  type="text"
                  value={notes}
                  onChange={e => setNotes(e.target.value)}
                  placeholder="e.g. Living room wall mount, static IP assigned"
                  className="w-full text-xs p-1.5 rounded bg-neutral-50 dark:bg-neutral-900 border border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100"
                />
              </div>
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleSaveEdit}
                className="px-3 py-1 bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-950 text-xs font-medium rounded flex items-center gap-1"
              >
                <Check className="w-3 h-3" />
                Save Changes
              </button>
              <button
                onClick={() => setIsEditing(false)}
                className="px-3 py-1 text-xs text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded"
              >
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Summary KPI Cards (Section 14) */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] font-semibold text-neutral-500 uppercase tracking-wider mb-0.5">
            Status
          </div>
          <div className="text-sm font-bold text-neutral-900 dark:text-neutral-100 capitalize">
            {selectedDevice.status}
          </div>
          <div className="text-[10px] text-neutral-400 mt-0.5">
            {selectedDevice.status === 'online' ? 'Responding to ping' : 'Host unreachable'}
          </div>
        </div>

        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] font-semibold text-neutral-500 uppercase tracking-wider mb-0.5">
            Latency
          </div>
          <div className="text-sm font-bold text-neutral-900 dark:text-neutral-100 font-mono tabular-nums">
            {selectedDevice.latencyMs !== undefined ? `${selectedDevice.latencyMs} ms` : '—'}
          </div>
          <div className="text-[10px] text-neutral-400 mt-0.5">ICMP round trip time</div>
        </div>

        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] font-semibold text-neutral-500 uppercase tracking-wider mb-0.5">
            First Seen
          </div>
          <div className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
            {selectedDevice.firstSeen}
          </div>
          <div className="text-[10px] text-neutral-400 mt-0.5">Initial ARP registration</div>
        </div>

        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] font-semibold text-neutral-500 uppercase tracking-wider mb-0.5">
            Last Seen
          </div>
          <div className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
            {selectedDevice.lastSeen}
          </div>
          <div className="text-[10px] text-neutral-400 mt-0.5">Recent heartbeat</div>
        </div>
      </div>

      {/* Two columns: Device Information & Detected Services */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Device Information Card */}
        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-4 space-y-3 transition-colors">
          <div className="flex items-center gap-1.5 text-xs font-bold text-neutral-900 dark:text-neutral-100 border-b border-neutral-100 dark:border-neutral-800 pb-2">
            <Shield className="w-3.5 h-3.5 text-neutral-500" />
            <span>Network Identification</span>
          </div>

          <div className="space-y-2 text-xs">
            <div className="flex justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
              <span className="text-neutral-500">IP Address</span>
              <span className="font-mono font-semibold text-neutral-900 dark:text-neutral-100">
                {selectedDevice.ip}
              </span>
            </div>

            <div className="flex justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
              <span className="text-neutral-500">MAC Address</span>
              <span className="font-mono text-neutral-800 dark:text-neutral-200">
                {selectedDevice.mac}
              </span>
            </div>

            <div className="flex justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
              <span className="text-neutral-500">Hostname</span>
              <span className="font-mono text-neutral-800 dark:text-neutral-200">
                {selectedDevice.hostname}
              </span>
            </div>

            <div className="flex justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
              <span className="text-neutral-500">Vendor</span>
              <span className="font-medium text-neutral-800 dark:text-neutral-200">
                {selectedDevice.vendor}
              </span>
            </div>

            <div className="flex justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
              <span className="text-neutral-500">Device Type</span>
              <span className="font-medium text-neutral-800 dark:text-neutral-200">
                {selectedDevice.type}
              </span>
            </div>

            {selectedDevice.os && (
              <div className="flex justify-between py-1 border-b border-neutral-100 dark:border-neutral-800/60">
                <span className="text-neutral-500">OS / Firmware</span>
                <span className="font-medium text-neutral-800 dark:text-neutral-200">
                  {selectedDevice.os}
                </span>
              </div>
            )}

            {selectedDevice.notes && (
              <div className="flex justify-between py-1">
                <span className="text-neutral-500">Notes</span>
                <span className="text-neutral-700 dark:text-neutral-300 italic">
                  {selectedDevice.notes}
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Device Services (Section 15) */}
        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-4 space-y-3 transition-colors">
          <div className="flex items-center justify-between border-b border-neutral-100 dark:border-neutral-800 pb-2">
            <div className="flex items-center gap-1.5 text-xs font-bold text-neutral-900 dark:text-neutral-100">
              <Layers className="w-3.5 h-3.5 text-neutral-500" />
              <span>Detected Services</span>
            </div>
            <button
              onClick={handleScanServices}
              disabled={isScanningPorts || selectedDevice.status === 'offline'}
              className="text-[11px] text-emerald-600 dark:text-emerald-400 font-medium hover:underline disabled:opacity-40"
            >
              {isScanningPorts ? 'Scanning...' : 'Rescan Ports'}
            </button>
          </div>

          {!selectedDevice.services || selectedDevice.services.length === 0 ? (
            <div className="py-8 text-center text-xs text-neutral-400">
              <p>No open services detected on this device.</p>
              <button
                onClick={handleScanServices}
                disabled={isScanningPorts || selectedDevice.status === 'offline'}
                className="mt-2 text-xs font-medium text-neutral-700 dark:text-neutral-300 underline"
              >
                Run service discovery scan
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="text-neutral-400 font-semibold uppercase text-[10px] border-b border-neutral-100 dark:border-neutral-800">
                    <th className="pb-1.5">Port</th>
                    <th className="pb-1.5">Proto</th>
                    <th className="pb-1.5">Service</th>
                    <th className="pb-1.5 text-right">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800">
                  {selectedDevice.services.map((svc, idx) => (
                    <tr key={idx} className="hover:bg-neutral-50 dark:hover:bg-neutral-800/50">
                      <td className="py-2 font-mono font-medium text-neutral-900 dark:text-neutral-100 tabular-nums">
                        {svc.port}
                      </td>
                      <td className="py-2 font-mono text-[11px] text-neutral-500">
                        {svc.protocol}
                      </td>
                      <td className="py-2 text-neutral-700 dark:text-neutral-300 font-medium">
                        {svc.service}
                      </td>
                      <td className="py-2 text-right">
                        <span className="inline-flex items-center gap-1 text-[11px] text-emerald-600 dark:text-emerald-400 font-medium">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                          {svc.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Device History (Section 16) */}
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-4 space-y-3 transition-colors">
        <div className="flex items-center gap-1.5 text-xs font-bold text-neutral-900 dark:text-neutral-100 border-b border-neutral-100 dark:border-neutral-800 pb-2">
          <Clock className="w-3.5 h-3.5 text-neutral-500" />
          <span>Device History</span>
        </div>

        <div className="space-y-3 pt-1">
          {(!selectedDevice.history || selectedDevice.history.length === 0) ? (
            <p className="text-xs text-neutral-400">No previous events recorded for this device.</p>
          ) : (
            selectedDevice.history.map(item => (
              <div key={item.id} className="flex items-start gap-3 text-xs">
                <span className="font-mono text-neutral-400 tabular-nums text-[11px] w-28 shrink-0">
                  {item.timestamp}
                </span>
                <span className={`w-1.5 h-1.5 rounded-full mt-1.5 shrink-0 ${
                  item.type === 'discovered' ? 'bg-amber-500' :
                  item.type === 'online' ? 'bg-emerald-500' :
                  item.type === 'offline' ? 'bg-neutral-400' : 'bg-sky-500'
                }`} />
                <span className="text-neutral-700 dark:text-neutral-300">
                  {item.description}
                </span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};
