import { Device, DeviceEvent, DeviceService } from '../types/device';
import { NetworkInfo, ScanResult } from '../types/network';
import { NetworkEvent } from '../types/events';
import { AppSettings, DatabaseStats, MaintenanceResult, PruneResult } from '../types/settings';

export type ScanProgressCallback = (progress: number, scanned: number, total: number, found: number) => void;

export interface NetworkService {
  /** true for the in-memory prototype; false when backed by the real engine. */
  readonly isSimulated: boolean;
  getNetworkInfo(): Promise<NetworkInfo>;
  getDevices(): Promise<Device[]>;
  getDevice(id: string): Promise<Device | undefined>;
  scanNetwork(type?: 'quick' | 'full', onProgress?: ScanProgressCallback): Promise<ScanResult>;
  getEvents(): Promise<NetworkEvent[]>;
  getDeviceHistory(id: string): Promise<DeviceEvent[]>;
  updateDevice(id: string, updates: Partial<Device>): Promise<Device>;
  pingDevice(ip: string): Promise<{ success: boolean; latencyMs: number }>;
  wakeOnLan(mac: string): Promise<{ success: boolean; message: string }>;
  scanDevicePorts(id: string): Promise<DeviceService[]>;
  mergeDevices?(targetId: string, sourceId: string): Promise<void>;
  
  // Settings & Database Management
  getSettings?(): Promise<AppSettings>;
  updateSettings?(settings: Partial<AppSettings>): Promise<AppSettings>;
  getDatabaseStats?(): Promise<DatabaseStats>;
  vacuumDatabase?(): Promise<MaintenanceResult>;
  integrityCheck?(): Promise<MaintenanceResult>;
  pruneEvents?(olderThanDays?: number): Promise<PruneResult>;

  // Optional host actions (real engine only)
  clearHistory?(): Promise<void>;
  openDataFolder?(): Promise<void>;

  // Test/Prototype simulation helpers
  setSimulatedEmpty(empty: boolean): void;
  setSimulatedError(error: boolean): void;
  toggleDeviceStatus(id: string): Promise<Device>;
  addNewSimulatedDevice(): Promise<Device>;
  resetToDefault(): Promise<void>;
}
