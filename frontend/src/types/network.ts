export type InterfaceStatus = 'Connected' | 'Disconnected';
export type InterfaceType = 'Wi-Fi' | 'Ethernet' | 'VPN';

export interface NetworkInterfaceItem {
  id: string;
  name: string;
  type: InterfaceType;
  adapterName: string;
  ip: string;
  subnet: string;
  gateway: string;
  mac: string;
  status: InterfaceStatus;
  speedMbps?: number;
  isDefault: boolean;
}

export interface NetworkPingStats {
  currentMs: number;
  avgMs: number;
  peakMs: number;
  packetLossPercent: number;
  history: number[];
}

export interface NetworkInfo {
  networkName: string;
  ssid: string;
  interfaceName: string;
  interfaceType: InterfaceType;
  gateway: string;
  subnet: string;
  dns: string[];
  localIp: string;
  publicIp?: string;
  isPublicNetwork?: boolean;
  broadcast: string;
  netmask: string;
  totalAddresses: number;
  activeAddresses: number;
  status: 'Connected' | 'Disconnected' | 'Error';
  pingStats: NetworkPingStats;
  interfaces: NetworkInterfaceItem[];
}

export interface ScanResult {
  scanId: string;
  type: 'quick' | 'full';
  scannedAddresses: number;
  totalAddresses: number;
  devicesFound: number;
  newDevices: number;
  errors: number;
  durationMs: number;
  timestamp: string;
}
