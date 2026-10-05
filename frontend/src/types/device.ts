export type DeviceType =
  | 'Router'
  | 'Computer'
  | 'Phone'
  | 'Tablet'
  | 'TV'
  | 'Printer'
  | 'Camera'
  | 'IoT'
  | 'Server'
  | 'Network Device'
  | 'Game Console'
  | 'Unknown';

export type DeviceStatus = 'online' | 'offline';

export interface DeviceService {
  port: number;
  protocol: 'TCP' | 'UDP';
  service: string;
  status: 'Open' | 'Closed' | 'Filtered';
}

export interface DeviceEvent {
  id: string;
  timestamp: string;
  type: 'discovered' | 'online' | 'offline' | 'ip_changed' | 'service_detected';
  description: string;
}

export interface Device {
  id: string;
  name: string;
  customAlias?: string;
  hostname: string;
  ip: string;
  mac: string;
  vendor: string;
  type: DeviceType;
  status: DeviceStatus;
  isNew?: boolean;
  firstSeen: string;
  lastSeen: string;
  /** RFC 3339 instant behind `lastSeen` (real engine only), used for sorting. */
  lastSeenAt?: string;
  latencyMs?: number;
  os?: string;
  notes?: string;
  services?: DeviceService[];
  history?: DeviceEvent[];
}
