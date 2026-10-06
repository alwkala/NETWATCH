import React, { useState } from 'react';
import { useNetwork } from '../context/NetworkContext';
import { DeviceTypeIcon } from '../components/devices/DeviceTypeIcon';
import {
  Wifi,
  Globe,
  Router,
  Layers,
  CheckCircle2,
  XCircle,
  Network as NetworkIcon,
  Shield,
  ArrowUpRight
} from 'lucide-react';

export const Network: React.FC = () => {
  const { networkInfo, devices, selectDevice } = useNetwork();
  // The allocation grid is drawn for the /24 containing this host.
  const subnetPrefix = (networkInfo?.localIp || '192.168.1.24').split('.').slice(0, 3).join('.') + '.';
  const [selectedSubnetTab, setSelectedSubnetTab] = useState<'topology' | 'interfaces' | 'allocation'>('topology');

  // Key devices for clean topology map
  const realRouter = devices.find(d => (networkInfo?.gateway && d.ip === networkInfo.gateway) || d.type === 'Router');
  const routerDevice = realRouter || {
    id: '',
    name: 'Default Gateway (Router)',
    ip: networkInfo?.gateway || '192.168.1.1',
    mac: '—',
    vendor: 'Gateway',
    type: 'Router' as const
  };

  const switchDevice = devices.find(d => d.type === 'Network Device');
  const onlineDevices = devices.filter(d => d.status === 'online' && d.id !== routerDevice.id && d.id !== switchDevice?.id);

  return (
    <div className="p-6 space-y-6 max-w-6xl mx-auto">
      {/* Header Overview */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-3 border-b border-neutral-200 dark:border-neutral-800">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-bold text-neutral-900 dark:text-neutral-100">
              {networkInfo?.networkName || 'Home Network'}
            </h2>
            <span className="font-mono text-xs px-2 py-0.5 rounded bg-emerald-50 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800 font-medium">
              ● Connected
            </span>
          </div>
          <p className="text-xs text-neutral-500 dark:text-neutral-400 mt-0.5">
            Active interface: <strong className="text-neutral-700 dark:text-neutral-300">{networkInfo?.interfaceName}</strong> ({networkInfo?.ssid}) · Subnet: <strong className="font-mono">{networkInfo?.subnet}</strong>
          </p>
        </div>

        {/* View mode toggle */}
        <div className="flex items-center gap-1 bg-neutral-100 dark:bg-neutral-800 p-1 rounded border border-neutral-200 dark:border-neutral-700">
          <button
            onClick={() => setSelectedSubnetTab('topology')}
            className={`px-3 py-1 text-xs font-medium rounded transition-colors ${
              selectedSubnetTab === 'topology'
                ? 'bg-white dark:bg-neutral-700 text-neutral-900 dark:text-neutral-100 shadow-xs'
                : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-200'
            }`}
          >
            Topology Map
          </button>
          <button
            onClick={() => setSelectedSubnetTab('interfaces')}
            className={`px-3 py-1 text-xs font-medium rounded transition-colors ${
              selectedSubnetTab === 'interfaces'
                ? 'bg-white dark:bg-neutral-700 text-neutral-900 dark:text-neutral-100 shadow-xs'
                : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-200'
            }`}
          >
            Network Interfaces
          </button>
          <button
            onClick={() => setSelectedSubnetTab('allocation')}
            className={`px-3 py-1 text-xs font-medium rounded transition-colors ${
              selectedSubnetTab === 'allocation'
                ? 'bg-white dark:bg-neutral-700 text-neutral-900 dark:text-neutral-100 shadow-xs'
                : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-200'
            }`}
          >
            Subnet Allocation
          </button>
        </div>
      </div>

      {/* Network Specs Cards (Section 17) */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-3 text-xs">
        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] uppercase font-semibold text-neutral-400 mb-0.5">Interface</div>
          <div className="font-semibold text-neutral-900 dark:text-neutral-100 flex items-center gap-1.5">
            <Wifi className="w-3.5 h-3.5 text-emerald-500" />
            <span>{networkInfo?.interfaceName}</span>
          </div>
          <div className="text-[11px] text-neutral-500">{networkInfo?.ssid}</div>
        </div>

        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] uppercase font-semibold text-neutral-400 mb-0.5">Local Host IP</div>
          <div className="font-mono font-semibold text-neutral-900 dark:text-neutral-100 tabular-nums">
            {networkInfo?.localIp}
          </div>
          <div className="text-[11px] text-neutral-500 font-mono">This device</div>
        </div>

        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] uppercase font-semibold text-neutral-400 mb-0.5">Gateway Router</div>
          <div className="font-mono font-semibold text-neutral-900 dark:text-neutral-100 tabular-nums">
            {networkInfo?.gateway}
          </div>
          <div className="text-[11px] text-neutral-500">router.local</div>
        </div>

        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] uppercase font-semibold text-neutral-400 mb-0.5">Subnet CIDR</div>
          <div className="font-mono font-semibold text-neutral-900 dark:text-neutral-100 tabular-nums">
            {networkInfo?.subnet}
          </div>
          <div className="text-[11px] text-neutral-500 font-mono">255.255.255.0</div>
        </div>

        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] uppercase font-semibold text-neutral-400 mb-0.5">DNS Servers</div>
          <div className="font-mono font-semibold text-neutral-900 dark:text-neutral-100 tabular-nums truncate">
            {networkInfo?.dns.join(', ')}
          </div>
          <div className="text-[11px] text-neutral-500">Local + Cloudflare</div>
        </div>

        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md p-3 transition-colors">
          <div className="text-[10px] uppercase font-semibold text-neutral-400 mb-0.5">Broadcast IP</div>
          <div className="font-mono font-semibold text-neutral-900 dark:text-neutral-100 tabular-nums">
            {networkInfo?.broadcast}
          </div>
          <div className="text-[11px] text-neutral-500">254 usable IPs</div>
        </div>
      </div>

      {/* Main Tab Content */}
      {selectedSubnetTab === 'topology' && (
        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-6 space-y-6 shadow-xs transition-colors">
          <div className="flex items-center justify-between border-b border-neutral-100 dark:border-neutral-800 pb-3">
            <div>
              <h3 className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
                Network Topology Architecture (Section 18)
              </h3>
              <p className="text-xs text-neutral-500 dark:text-neutral-400">
                Conceptual local physical & logical hierarchy. Click any connected node to inspect.
              </p>
            </div>
            <span className="text-xs font-mono text-neutral-400">
              Subnet Layer 2/3
            </span>
          </div>

          {/* Technical Diagram Container */}
          <div className="py-4 flex flex-col items-center">
            {/* 1. INTERNET NODE */}
            <div className="flex flex-col items-center">
              <div className="px-5 py-2.5 rounded-md bg-neutral-100 dark:bg-neutral-800 border border-neutral-300 dark:border-neutral-700 flex items-center gap-2 text-xs font-semibold text-neutral-800 dark:text-neutral-200">
                <Globe className="w-4 h-4 text-sky-500" />
                <span>WAN / INTERNET</span>
              </div>
              {/* Connector line */}
              <div className="w-px h-8 bg-neutral-300 dark:bg-neutral-700" />
            </div>

            {/* 2. ROUTER NODE */}
            <div className="flex flex-col items-center">
              <button
                onClick={() => {
                  if (routerDevice.id) {
                    selectDevice(routerDevice.id);
                  }
                }}
                className={`px-5 py-3 rounded-md bg-white dark:bg-neutral-900 border-2 border-neutral-300 dark:border-neutral-700 ${
                  routerDevice.id ? 'hover:border-emerald-500 dark:hover:border-emerald-500 cursor-pointer' : 'cursor-default'
                } flex items-center gap-3 text-left transition-colors group shadow-xs`}
              >
                <div className="w-8 h-8 rounded bg-neutral-100 dark:bg-neutral-800 flex items-center justify-center text-neutral-700 dark:text-neutral-300">
                  <Router className="w-4 h-4 text-emerald-500" />
                </div>
                <div>
                  <div className="text-xs font-bold text-neutral-900 dark:text-neutral-100 group-hover:text-emerald-600 dark:group-hover:text-emerald-400">
                    {routerDevice.name}
                  </div>
                  <div className="text-[11px] font-mono text-neutral-500">
                    IP: {routerDevice.ip} {routerDevice.mac && routerDevice.mac !== '—' ? `· ${routerDevice.mac}` : ''}
                  </div>
                </div>
              </button>
              {/* Connector line */}
              <div className="w-px h-8 bg-neutral-300 dark:bg-neutral-700" />
            </div>

            {/* 3. SWITCH / ACCESS POINT TIER */}
            {switchDevice && (
              <div className="flex flex-col items-center">
                <button
                  onClick={() => selectDevice(switchDevice.id)}
                  className="px-4 py-2 rounded-md bg-neutral-50 dark:bg-neutral-800/80 border border-neutral-300 dark:border-neutral-700 hover:border-neutral-400 flex items-center gap-2 text-xs text-neutral-700 dark:text-neutral-300 transition-colors"
                >
                  <NetworkIcon className="w-3.5 h-3.5 text-neutral-500" />
                  <span className="font-semibold">{switchDevice.name}</span>
                  <span className="font-mono text-[11px] text-neutral-400">({switchDevice.ip})</span>
                </button>
                <div className="w-px h-6 bg-neutral-300 dark:bg-neutral-700" />
              </div>
            )}

            {/* 4. BUS / FAN-OUT TO CONNECTED DEVICES */}
            <div className="w-full max-w-4xl">
              {/* Horizontal crossbar */}
              <div className="h-px bg-neutral-300 dark:bg-neutral-700 w-full mb-6 relative">
                <span className="absolute -top-2.5 left-1/2 -translate-x-1/2 px-2 bg-white dark:bg-neutral-900 text-[10px] font-mono text-neutral-400">
                  {networkInfo?.subnet || '192.168.1.0/24'} Broadcast Domain
                </span>
              </div>

              {/* Grid of online connected devices */}
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2.5">
                {onlineDevices.slice(0, 12).map(dev => (
                  <button
                    key={dev.id}
                    onClick={() => selectDevice(dev.id)}
                    className="p-2.5 rounded-md bg-neutral-50 dark:bg-neutral-800/60 border border-neutral-200 dark:border-neutral-700/80 hover:border-emerald-500 hover:bg-neutral-100 dark:hover:bg-neutral-800 text-left transition-all group flex flex-col justify-between"
                  >
                    <div className="flex items-center justify-between mb-1.5">
                      <DeviceTypeIcon
                        type={dev.type}
                        className="w-3.5 h-3.5 text-neutral-500 group-hover:text-emerald-500"
                      />
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                    </div>
                    <div className="text-xs font-semibold text-neutral-900 dark:text-neutral-100 truncate group-hover:text-emerald-600 dark:group-hover:text-emerald-400">
                      {dev.customAlias || dev.name}
                    </div>
                    <div className="text-[11px] font-mono text-neutral-500 truncate tabular-nums mt-0.5">
                      {dev.ip}
                    </div>
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Network Interfaces Tab (Section 17) */}
      {selectedSubnetTab === 'interfaces' && (
        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs transition-colors">
          <div className="border-b border-neutral-100 dark:border-neutral-800 pb-2">
            <h3 className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
              Host Network Adapters & Interfaces
            </h3>
            <p className="text-xs text-neutral-500 dark:text-neutral-400">
              Physical and virtual network interfaces detected on this Windows workstation.
            </p>
          </div>

          <div className="space-y-3">
            {networkInfo?.interfaces.map(iface => {
              const isConn = iface.status === 'Connected';
              return (
                <div
                  key={iface.id}
                  className={`p-4 rounded-md border text-xs flex flex-col sm:flex-row sm:items-center justify-between gap-3 ${
                    isConn
                      ? 'bg-neutral-50 dark:bg-neutral-800/60 border-neutral-300 dark:border-neutral-700'
                      : 'bg-neutral-50/50 dark:bg-neutral-900/40 border-neutral-200 dark:border-neutral-800 opacity-60'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div className="w-9 h-9 rounded bg-white dark:bg-neutral-800 flex items-center justify-center text-neutral-700 dark:text-neutral-300 border border-neutral-200 dark:border-neutral-700">
                      <NetworkIcon className="w-4 h-4" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-neutral-900 dark:text-neutral-100 text-sm">
                          {iface.name}
                        </span>
                        {iface.isDefault && (
                          <span className="text-[10px] font-mono px-1.5 py-0.2 bg-neutral-200 dark:bg-neutral-700 text-neutral-700 dark:text-neutral-300 rounded font-semibold">
                            DEFAULT GATEWAY
                          </span>
                        )}
                      </div>
                      <div className="text-neutral-500 text-[11px] mt-0.5">
                        {iface.adapterName}
                      </div>
                      <div className="text-neutral-400 font-mono text-[11px]">
                        MAC: {iface.mac}
                      </div>
                    </div>
                  </div>

                  <div className="sm:text-right font-mono space-y-0.5">
                    <div className="flex sm:justify-end items-center gap-1.5 font-sans font-medium text-xs">
                      {isConn ? (
                        <span className="text-emerald-600 dark:text-emerald-400 flex items-center gap-1">
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          Connected
                        </span>
                      ) : (
                        <span className="text-neutral-400 flex items-center gap-1">
                          <XCircle className="w-3.5 h-3.5" />
                          Disconnected
                        </span>
                      )}
                    </div>
                    {isConn && (
                      <>
                        <div className="font-semibold text-neutral-900 dark:text-neutral-100">
                          {iface.ip}
                        </div>
                        <div className="text-[11px] text-neutral-400">
                          Subnet: {iface.subnet}
                        </div>
                        {iface.speedMbps && (
                          <div className="text-[10px] text-neutral-500 font-sans">
                            Link Speed: {iface.speedMbps} Mbps
                          </div>
                        )}
                      </>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Subnet Allocation Grid */}
      {selectedSubnetTab === 'allocation' && (
        <div className="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-5 space-y-4 shadow-xs transition-colors">
          <div className="border-b border-neutral-100 dark:border-neutral-800 pb-2">
            <h3 className="text-sm font-bold text-neutral-900 dark:text-neutral-100">
              IPv4 Subnet Address Space ({subnetPrefix}1 — {subnetPrefix}254)
            </h3>
            <p className="text-xs text-neutral-500 dark:text-neutral-400">
              Visual map of assigned IP leases vs available unassigned addresses in /24 subnet.
            </p>
          </div>

          {/* Subnet heat grid (254 cells) */}
          <div className="grid grid-cols-16 sm:grid-cols-32 gap-1 p-3 bg-neutral-50 dark:bg-neutral-900 rounded border border-neutral-200 dark:border-neutral-800">
            {Array.from({ length: 254 }, (_, i) => {
              const octet = i + 1;
              const ipStr = `${subnetPrefix}${octet}`;
              const matchedDevice = devices.find(d => d.ip === ipStr);
              const isLocalHost = networkInfo?.localIp === ipStr;
              const isGateway = networkInfo?.gateway === ipStr;

              return (
                <div
                  key={octet}
                  onClick={() => matchedDevice && selectDevice(matchedDevice.id)}
                  title={`${ipStr}${matchedDevice ? ` - ${matchedDevice.name} (${matchedDevice.status})` : ' (Available)'}`}
                  className={`h-4 rounded-xs transition-colors cursor-pointer ${
                    isGateway
                      ? 'bg-amber-500'
                      : isLocalHost
                      ? 'bg-sky-500'
                      : matchedDevice?.status === 'online'
                      ? 'bg-emerald-500'
                      : matchedDevice?.status === 'offline'
                      ? 'bg-neutral-400 dark:bg-neutral-600'
                      : 'bg-neutral-200 dark:bg-neutral-800 hover:bg-neutral-300 dark:hover:bg-neutral-700'
                  }`}
                />
              );
            })}
          </div>

          {/* Legend */}
          <div className="flex flex-wrap items-center gap-4 text-xs text-neutral-600 dark:text-neutral-400 pt-1 font-mono text-[11px]">
            <div className="flex items-center gap-1.5">
              <span className="w-3 h-3 rounded-xs bg-amber-500" />
              <span>Gateway (.1)</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-3 h-3 rounded-xs bg-sky-500" />
              <span>This PC (.24)</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-3 h-3 rounded-xs bg-emerald-500" />
              <span>Active Hosts ({devices.filter(d => d.status === 'online').length})</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-3 h-3 rounded-xs bg-neutral-400 dark:bg-neutral-600" />
              <span>Offline Cache ({devices.filter(d => d.status === 'offline').length})</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-3 h-3 rounded-xs bg-neutral-200 dark:bg-neutral-800" />
              <span>Available (230)</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
