export interface AppSettings {
  autoDiscovery: boolean;
  scanInterval: string; // '1m' | '5m' | '15m' | '1h' | 'manual'
  notifyNewDevice: boolean;
  notifyDeviceOffline: boolean;
  notifyNetworkChange: boolean;
  launchAtStartup: boolean;
  startMinimized: boolean;
}

export interface DatabaseStats {
  dbPath: string;
  fileSizeBytes: number;
  deviceCount: number;
  eventCount: number;
  scanCount: number;
  walEnabled: boolean;
}

export interface MaintenanceResult {
  success: boolean;
  message: string;
}

export interface PruneResult {
  deletedCount: number;
  olderThanDays: number;
}
