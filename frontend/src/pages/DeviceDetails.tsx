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
  Power,
  GitMerge,
  UserCheck,
  Copy,
  AlertTriangle
} from 'lucide-react';
import { TrustStatus, DeviceType } from '../types/device';

const DEVICE_TYPE_OPTIONS: { type: DeviceType; label: string }[] = [
  { type: 'Computer', label: 'Computer / PC' },
  { type: 'Phone', label: 'Phone / Mobile' },
  { type: 'Tablet', label: 'Tablet' },
  { type: 'Router', label: 'Router / Gateway' },
  { type: 'Server', label: 'Server / NAS' },
  { type: 'Network Device', label: 'Switch / AP / Network' },
  { type: 'Printer', label: 'Printer' },
  { type: 'TV', label: 'Smart TV / Display' },
  { type: 'Camera', label: 'IP Camera' },
  { type: 'Game Console', label: 'Game Console' },
  { type: 'IoT', label: 'IoT / Smart Device' },
  { type: 'Unknown', label: 'Unknown' },
];

export const DeviceDetails: React.FC = () => {
  const {
    selectedDevice,
    devices,
    navigateTo,
    pingDevice,
    wakeOnLan,
    scanDevicePorts,
    updateDevice,
    mergeDevices,
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
  const [deviceType, setDeviceType] = useState<DeviceType>(selectedDevice?.type || 'Unknown');
  const [notes, setNotes] = useState('');
  const [trustStatus, setTrustStatus] = useState<TrustStatus>('unknown');
  const [isMerging, setIsMerging] = useState(false);
  const [targetMergeId, setTargetMergeId] = useState('');
  const [isMergeSubmitting, setIsMergeSubmitting] = useState(false);
  const [copiedLocation, setCopiedLocation] = useState(false);

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

  const evidence = selectedDevice.evidence || [];
  const mdnsItems = evidence.filter(e => e.source === 'mDNS');
  const ssdpItems = evidence.filter(e => e.source === 'SSDP');
  const nbnsItems = evidence.filter(e => e.source === 'NBNS');

  const mdnsHost = mdnsItems.find(e => e.key === 'hostname')?.value;
  const mdnsServices = mdnsItems.filter(e => e.key === 'service').map(e => e.value);

  const ssdpST = ssdpItems.find(e => e.key === 'st')?.value;
  const ssdpServer = ssdpItems.find(e => e.key === 'server')?.value;
  const ssdpLocation = ssdpItems.find(e => e.key === 'location')?.value;

  const nbnsName = nbnsItems.find(e => e.key === 'hostname')?.value;
  const nbnsUnitID = nbnsItems.find(e => e.key === 'unit_id')?.value;
  const nbnsMACMatch = nbnsItems.find(e => e.key === 'mac_match')?.value;

  const primaryTitle = selectedDevice.customAlias || selectedDevice.name;
  const otherNames = Array.from(
    new Set(
      evidence
        .filter(e => e.key === 'hostname' && e.value && e.value !== primaryTitle)
        .map(e => e.value)
    )
  );

  const handleCopyLocation = (url: string) => {
    navigator.clipboard.writeText(url);
    setCopiedLocation(true);
    setTimeout(() => setCopiedLocation(false), 2000);
  };

  const handleStartEdit = () => {
    setCustomAlias(selectedDevice.customAlias || selectedDevice.name);
    setNotes(selectedDevice.notes || '');
    setTrustStatus(selectedDevice.trustStatus || 'unknown');
    setDeviceType(selectedDevice.type || 'Unknown');
    setIsEditing(true);
  };

  const handleSaveEdit = async () => {
    await updateDevice(selectedDevice.id, {
      customAlias: customAlias.trim() || undefined,
      notes: notes.trim() || undefined,
      trustStatus,
      type: deviceType,
    });
    setIsEditing(false);
  };

  const handleQuickTrustChange = async (newTrust: TrustStatus) => {
    await updateDevice(selectedDevice.id, { trustStatus: newTrust });
  };

  const handleConfirmMerge = async () => {
    if (!targetMergeId) return;
    setIsMergeSubmitting(true);
    try {
      await mergeDevices(targetMergeId, selectedDevice.id);
      setIsMerging(false);
    } finally {
      setIsMergeSubmitting(false);
    }
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
              <div className="flex items-center gap-2 flex-wrap">
                <h2 className="text-xl font-bold text-neutral-900 dark:text-neutral-100">
                  {selectedDevice.customAlias || selectedDevice.name}
                </h2>
                <DeviceStatusBadge
                  status={selectedDevice.status}
                  isNew={selectedDevice.isNew}
                />
                {selectedDevice.isRandomizedMac && (
                  <span
                    className="px-2 py-0.5 rounded text-[11px] font-mono bg-purple-50 dark:bg-purple-950/60 text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800"
                    title="Locally Administered MAC (Private address generated by iOS/Android/Windows Private Wi-Fi)"
                  >
                    Private MAC
                  </span>
                )}
                {/* Trust status pill */}
                <div className="inline-flex items-center gap-1 bg-neutral-100 dark:bg-neutral-800 p-0.5 rounded border border-neutral-200 dark:border-neutral-700 text-[10px]">
                  {(['known', 'guest', 'unknown'] as const).map(t => (
                    <button
                      key={t}
                      onClick={() => handleQuickTrustChange(t)}
                      className={`px-2 py-0.5 rounded capitalize font-medium transition-colors ${
                        (selectedDevice.trustStatus || 'unknown') === t
                          ? t === 'known'
                            ? 'bg-emerald-600 text-white'
                            : t === 'guest'
                            ? 'bg-blue-600 text-white'
                            : 'bg-neutral-700 text-white dark:bg-neutral-600'
                          : 'text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-200'
                      }`}
                      title={`Set trust status to ${t}`}
                    >
                      {t}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex flex-wrap items-center gap-2 text-xs font-mono text-neutral-500 dark:text-neutral-400 mt-1">
                <span className="text-neutral-900 dark:text-neutral-200 font-semibold">{selectedDevice.ip}</span>
                <span>·</span>
                <span>{selectedDevice.mac}</span>
                <span>·</span>
                <span className="font-sans text-neutral-600 dark:text-neutral-400">{selectedDevice.vendor}</span>
              </div>

              {otherNames.length > 0 && (
                <div className="flex flex-wrap items-center gap-1.5 text-[11px] text-neutral-500 mt-1">
                  <span>Other observed names:</span>
                  {otherNames.map((n, i) => (
                    <span key={i} className="font-mono text-neutral-700 dark:text-neutral-300 bg-neutral-100 dark:bg-neutral-800 px-1.5 py-0.5 rounded text-[10px] border border-neutral-200/60 dark:border-neutral-700/60">
                      {n}
                    </span>
                  ))}
                </div>
              )}
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

            <button
              onClick={() => setIsMerging(true)}
              className="px-3 py-1.5 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-800 dark:text-neutral-200 rounded text-xs font-medium border border-neutral-300 dark:border-neutral-700 flex items-center gap-1.5 transition-colors"
              title="Merge this device into another canonical record (e.g. for private Wi-Fi MACs)"
            >
              <GitMerge className="w-3.5 h-3.5 text-purple-500" />
              <span>Merge Device</span>
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
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
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
                  Device Type & Icon
                </label>
                <div className="flex items-center gap-1.5">
                  <div className="w-7 h-7 shrink-0 rounded bg-neutral-100 dark:bg-neutral-800 flex items-center justify-center text-neutral-700 dark:text-neutral-300 border border-neutral-300 dark:border-neutral-700 shadow-2xs">
                    <DeviceTypeIcon type={deviceType} className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
                  </div>
                  <select
                    value={deviceType}
                    onChange={e => setDeviceType(e.target.value as DeviceType)}
                    className="w-full text-xs p-1.5 rounded bg-neutral-50 dark:bg-neutral-900 border border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100 font-medium"
                  >
                    {DEVICE_TYPE_OPTIONS.map(opt => (
                      <option key={opt.type} value={opt.type}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <div>
                <label className="text-[11px] text-neutral-500 block mb-1">
                  Trust Status
                </label>
                <select
                  value={trustStatus}
                  onChange={e => setTrustStatus(e.target.value as TrustStatus)}
                  className="w-full text-xs p-1.5 rounded bg-neutral-50 dark:bg-neutral-900 border border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100"
                >
                  <option value="known">Known (Approved Asset)</option>
                  <option value="guest">Guest (Temporary Visitor)</option>
                  <option value="unknown">Unknown (Unclassified)</option>
                </select>
              </div>
              <div>
                <label className="text-[11px] text-neutral-500 block mb-1">
                  Local Notes
                </label>
                <input
                  type="text"
                  value={notes}
                  onChange={e => setNotes(e.target.value)}
                  placeholder="e.g. Living room wall mount, static IP"
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

        {/* Device Merge Drawer */}
        {isMerging && (
          <div className="mt-4 pt-4 border-t border-purple-200 dark:border-purple-900/60 bg-purple-50/50 dark:bg-purple-950/20 p-3 rounded space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5 text-xs font-bold text-purple-900 dark:text-purple-300">
                <GitMerge className="w-4 h-4 text-purple-600 dark:text-purple-400" />
                <span>Merge Device Record</span>
              </div>
              <button onClick={() => setIsMerging(false)} className="text-neutral-400 hover:text-neutral-600">
                <X className="w-3.5 h-3.5" />
              </button>
            </div>
            <p className="text-[11px] text-neutral-600 dark:text-neutral-400">
              Unify this device record ({selectedDevice.customAlias || selectedDevice.name} · {selectedDevice.mac}) into a canonical target device.
              Useful for phones and laptops with rotating private MAC addresses.
              All discovery history will be combined, this MAC will be registered as an alias, and this duplicate entry will be removed.
            </p>
            <div>
              <label className="text-[11px] font-semibold text-neutral-700 dark:text-neutral-300 block mb-1">
                Select Target Canonical Device:
              </label>
              <select
                value={targetMergeId}
                onChange={e => setTargetMergeId(e.target.value)}
                className="w-full text-xs p-1.5 rounded bg-white dark:bg-neutral-900 border border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100"
              >
                <option value="">-- Choose target device --</option>
                {devices
                  .filter(d => d.id !== selectedDevice.id)
                  .map(d => (
                    <option key={d.id} value={d.id}>
                      {d.customAlias || d.name} ({d.ip} · {d.mac})
                    </option>
                  ))}
              </select>
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleConfirmMerge}
                disabled={!targetMergeId || isMergeSubmitting}
                className="px-3 py-1.5 bg-purple-600 hover:bg-purple-700 disabled:opacity-50 text-white text-xs font-medium rounded flex items-center gap-1"
              >
                {isMergeSubmitting ? <RotateCw className="w-3 h-3 animate-spin" /> : <GitMerge className="w-3 h-3" />}
                <span>Confirm Merge</span>
              </button>
              <button
                onClick={() => setIsMerging(false)}
                className="px-3 py-1.5 text-xs text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded"
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

      {/* Identification Evidence (Multi-Protocol Discovery) */}
      <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-4 space-y-4 transition-colors">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-neutral-100 dark:border-neutral-800 pb-2">
          <div className="flex items-center gap-1.5 text-xs font-bold text-neutral-900 dark:text-neutral-100">
            <Shield className="w-3.5 h-3.5 text-emerald-500" />
            <span>Identification Evidence</span>
            <span className="text-[10px] font-normal text-neutral-400">· Multi-Protocol Intelligence</span>
          </div>
          {/* Protocol Badges Summary */}
          <div className="flex items-center gap-1.5 flex-wrap">
            {['ARP', 'ICMP', 'NBNS', 'mDNS', 'SSDP'].map(proto => {
              const hasProto = proto === 'ARP' || proto === 'ICMP'
                ? true
                : (evidence.some(e => e.source === proto));
              return (
                <span
                  key={proto}
                  className={`text-[10px] font-mono px-1.5 py-0.5 rounded border transition-colors ${
                    hasProto
                      ? 'bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-800/80 font-medium'
                      : 'bg-neutral-50 dark:bg-neutral-800/50 text-neutral-400 dark:text-neutral-500 border-neutral-200/60 dark:border-neutral-700/60 opacity-60'
                  }`}
                >
                  {proto}
                </span>
              );
            })}
          </div>
        </div>

        {/* Structured Evidence Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
          {/* mDNS Section */}
          <div className="p-3 bg-neutral-50 dark:bg-neutral-800/40 rounded border border-neutral-100 dark:border-neutral-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-neutral-800 dark:text-neutral-200">Multicast DNS (mDNS)</span>
              <span className="text-[10px] font-mono text-neutral-400">224.0.0.251:5353</span>
            </div>
            {mdnsItems.length === 0 ? (
              <p className="text-[11px] text-neutral-400 italic">No mDNS records advertised</p>
            ) : (
              <div className="space-y-1.5 font-mono text-[11px]">
                {mdnsHost && (
                  <div className="flex justify-between items-center">
                    <span className="text-neutral-500 font-sans">Hostname:</span>
                    <span className="text-neutral-800 dark:text-neutral-200">{mdnsHost}</span>
                  </div>
                )}
                {mdnsServices.length > 0 && (
                  <div>
                    <span className="text-neutral-500 font-sans block mb-1">Advertised Services:</span>
                    <div className="flex flex-wrap gap-1">
                      {mdnsServices.map((s, i) => (
                        <span key={i} className="px-1.5 py-0.5 bg-neutral-200/70 dark:bg-neutral-700 text-neutral-800 dark:text-neutral-200 rounded text-[10px]">
                          {s}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* SSDP Section */}
          <div className="p-3 bg-neutral-50 dark:bg-neutral-800/40 rounded border border-neutral-100 dark:border-neutral-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-neutral-800 dark:text-neutral-200">SSDP / UPnP</span>
              <span className="text-[10px] font-mono text-neutral-400">239.255.255.250:1900</span>
            </div>
            {ssdpItems.length === 0 ? (
              <p className="text-[11px] text-neutral-400 italic">No UPnP announcements observed</p>
            ) : (
              <div className="space-y-1.5 font-mono text-[11px]">
                {ssdpST && (
                  <div className="flex justify-between gap-2">
                    <span className="text-neutral-500 font-sans shrink-0">Target:</span>
                    <span className="text-neutral-800 dark:text-neutral-200 truncate" title={ssdpST}>{ssdpST}</span>
                  </div>
                )}
                {ssdpServer && (
                  <div className="flex justify-between gap-2">
                    <span className="text-neutral-500 font-sans shrink-0">Server:</span>
                    <span className="text-neutral-800 dark:text-neutral-200 truncate" title={ssdpServer}>{ssdpServer}</span>
                  </div>
                )}
                {ssdpLocation && (
                  <div className="pt-1.5 border-t border-neutral-200/60 dark:border-neutral-700/60">
                    <div className="flex items-center justify-between text-[10px]">
                      <span className="text-neutral-500 font-sans">Location URL:</span>
                      <span className="px-1.5 py-0.5 rounded bg-amber-50 dark:bg-amber-950/40 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-800/60 font-sans text-[10px]">
                        Recorded — not fetched
                      </span>
                    </div>
                    <div className="flex items-center justify-between gap-2 mt-1 bg-white dark:bg-neutral-900 p-1.5 rounded border border-neutral-200/80 dark:border-neutral-700/80">
                      <span className="text-[11px] text-neutral-700 dark:text-neutral-300 truncate select-all">{ssdpLocation}</span>
                      <button
                        onClick={() => handleCopyLocation(ssdpLocation)}
                        className="text-[10px] text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-200 shrink-0 p-1 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded transition-colors"
                        title="Copy Location URL"
                      >
                        {copiedLocation ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* NBNS Section */}
          <div className="p-3 bg-neutral-50 dark:bg-neutral-800/40 rounded border border-neutral-100 dark:border-neutral-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-neutral-800 dark:text-neutral-200">NetBIOS Name Service</span>
              <span className="text-[10px] font-mono text-neutral-400">UDP 137</span>
            </div>
            {nbnsItems.length === 0 ? (
              <p className="text-[11px] text-neutral-400 italic">No NetBIOS response</p>
            ) : (
              <div className="space-y-1.5 font-mono text-[11px]">
                {nbnsName && (
                  <div className="flex justify-between items-center">
                    <span className="text-neutral-500 font-sans">Computer Name:</span>
                    <span className="text-neutral-800 dark:text-neutral-200">{nbnsName}</span>
                  </div>
                )}
                {nbnsUnitID && (
                  <div className="flex justify-between items-center">
                    <span className="text-neutral-500 font-sans">Unit ID (MAC):</span>
                    <div className="flex items-center gap-1.5">
                      <span className="text-neutral-800 dark:text-neutral-200">{nbnsUnitID}</span>
                      {nbnsMACMatch === 'true' && (
                        <span className="inline-flex items-center text-emerald-600 dark:text-emerald-400 text-[10px] gap-0.5 font-sans font-medium" title="Corroborates ARP MAC">
                          <Check className="w-3 h-3" /> Match
                        </span>
                      )}
                      {nbnsMACMatch === 'mismatch' && (
                        <span className="inline-flex items-center text-amber-600 dark:text-amber-400 text-[10px] gap-0.5 font-sans font-bold" title="Does not match ARP MAC">
                          <AlertTriangle className="w-3 h-3" /> Mismatch
                        </span>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* ARP / Link Layer Section */}
          <div className="p-3 bg-neutral-50 dark:bg-neutral-800/40 rounded border border-neutral-100 dark:border-neutral-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-neutral-800 dark:text-neutral-200">ARP & Layer 2</span>
              <span className="text-[10px] font-mono text-neutral-400">Neighbor Cache</span>
            </div>
            <div className="space-y-1.5 font-mono text-[11px]">
              <div className="flex justify-between">
                <span className="text-neutral-500 font-sans">Hardware MAC:</span>
                <span className="text-neutral-800 dark:text-neutral-200">{selectedDevice.mac}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-neutral-500 font-sans">IEEE Vendor:</span>
                <span className="text-neutral-800 dark:text-neutral-200 font-sans">{selectedDevice.vendor || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-neutral-500 font-sans">Inventory State:</span>
                <span className="text-neutral-800 dark:text-neutral-200 capitalize font-sans">{selectedDevice.status}</span>
              </div>
            </div>
          </div>
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
