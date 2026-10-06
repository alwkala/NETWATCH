import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { Device, DeviceService } from '../types/device';
import { NetworkInfo, ScanResult } from '../types/network';
import { NetworkEvent } from '../types/events';
import { NetworkService } from '../services/NetworkService';
import { defaultNetworkService } from '../services/MockNetworkService';

export type AppPage =
  | 'dashboard'
  | 'devices'
  | 'device-details'
  | 'network'
  | 'scanner'
  | 'events'
  | 'settings';

export type PrototypeStatePreset =
  | 'normal'
  | 'scanning'
  | 'scan_completed'
  | 'new_device'
  | 'device_offline'
  | 'empty'
  | 'loading'
  | 'error';

interface ScanProgressInfo {
  progress: number;
  scanned: number;
  total: number;
  found: number;
}

interface NetworkContextType {
  service: NetworkService;
  devices: Device[];
  networkInfo: NetworkInfo | null;
  events: NetworkEvent[];
  selectedDevice: Device | null;
  selectedDeviceId: string | null;
  activePage: AppPage;
  isScanning: boolean;
  scanProgress: ScanProgressInfo;
  lastScanResult: ScanResult | null;
  isLoading: boolean;
  error: string | null;
  prototypeState: PrototypeStatePreset;

  // Actions
  navigateTo: (page: AppPage, deviceId?: string) => void;
  selectDevice: (id: string | null) => void;
  startScan: (type?: 'quick' | 'full') => Promise<ScanResult | null>;
  refresh: () => Promise<void>;
  updateDevice: (id: string, updates: Partial<Device>) => Promise<void>;
  mergeDevices: (targetId: string, sourceId: string) => Promise<void>;
  pingDevice: (ip: string) => Promise<{ success: boolean; latencyMs: number }>;
  wakeOnLan: (mac: string) => Promise<{ success: boolean; message: string }>;
  scanDevicePorts: (id: string) => Promise<DeviceService[]>;
  toggleDeviceStatus: (id: string) => Promise<void>;
  addNewSimulatedDevice: () => Promise<void>;
  setPrototypeStatePreset: (preset: PrototypeStatePreset) => Promise<void>;
  clearError: () => void;
}

const NetworkContext = createContext<NetworkContextType | undefined>(undefined);

export const NetworkProvider: React.FC<{
  children: React.ReactNode;
  service?: NetworkService;
  startupError?: string;
}> = ({ children, service = defaultNetworkService, startupError }) => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [networkInfo, setNetworkInfo] = useState<NetworkInfo | null>(null);
  const [events, setEvents] = useState<NetworkEvent[]>([]);
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null);
  const [activePage, setActivePage] = useState<AppPage>('dashboard');
  const [isScanning, setIsScanning] = useState<boolean>(false);
  const [scanProgress, setScanProgress] = useState<ScanProgressInfo>({
    progress: 0,
    scanned: 0,
    total: 24,
    found: 0
  });
  const [lastScanResult, setLastScanResult] = useState<ScanResult | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(startupError ?? null);
  const [prototypeState, setPrototypeState] = useState<PrototypeStatePreset>('normal');

  const loadData = useCallback(async () => {
    try {
      setError(null);
      const [devs, net, evts] = await Promise.all([
        service.getDevices(),
        service.getNetworkInfo(),
        service.getEvents()
      ]);
      setDevices(devs);
      setNetworkInfo(net);
      setEvents(evts);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to query local network state';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  }, [service]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const selectDevice = useCallback((id: string | null) => {
    setSelectedDeviceId(id);
    if (id) {
      setActivePage('device-details');
    }
  }, []);

  const navigateTo = useCallback((page: AppPage, deviceId?: string) => {
    setActivePage(page);
    if (deviceId !== undefined) {
      setSelectedDeviceId(deviceId);
    } else if (page !== 'device-details') {
      setSelectedDeviceId(null);
    }
  }, []);

  const startScan = useCallback(
    async (type: 'quick' | 'full' = 'quick'): Promise<ScanResult | null> => {
      if (isScanning) return null;
      setIsScanning(true);
      setError(null);
      setScanProgress({ progress: 0, scanned: 0, total: 24, found: 0 });

      try {
        const result = await service.scanNetwork(type, (progress, scanned, total, found) => {
          setScanProgress({ progress, scanned, total, found });
        });
        setLastScanResult(result);
        const [devs, net, evts] = await Promise.all([
          service.getDevices(),
          service.getNetworkInfo(),
          service.getEvents()
        ]);
        setDevices(devs);
        setNetworkInfo(net);
        setEvents(evts);
        setPrototypeState('scan_completed');
        return result;
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : 'Network scan encountered an error';
        setError(message);
        return null;
      } finally {
        setIsScanning(false);
      }
    },
    [isScanning, service]
  );

  const refresh = useCallback(async () => {
    setIsLoading(true);
    await loadData();
  }, [loadData]);

  const updateDevice = useCallback(
    async (id: string, updates: Partial<Device>) => {
      try {
        const updated = await service.updateDevice(id, updates);
        setDevices(prev => prev.map(d => (d.id === id ? updated : d)));
      } catch (err: unknown) {
        console.error('Failed to update device', err);
      }
    },
    [service]
  );

  const mergeDevices = useCallback(
    async (targetId: string, sourceId: string) => {
      try {
        if (service.mergeDevices) {
          await service.mergeDevices(targetId, sourceId);
          await loadData();
          setSelectedDeviceId(targetId);
        }
      } catch (err: unknown) {
        console.error('Failed to merge devices', err);
        const message = err instanceof Error ? err.message : 'Failed to merge devices';
        setError(message);
      }
    },
    [service, loadData]
  );

  const pingDevice = useCallback(
    async (ip: string) => {
      return service.pingDevice(ip);
    },
    [service]
  );

  const wakeOnLan = useCallback(
    async (mac: string) => {
      return service.wakeOnLan(mac);
    },
    [service]
  );

  const scanDevicePorts = useCallback(
    async (id: string) => {
      const openPorts = await service.scanDevicePorts(id);
      setDevices(prev =>
        prev.map(d => (d.id === id ? { ...d, services: openPorts } : d))
      );
      return openPorts;
    },
    [service]
  );

  const toggleDeviceStatus = useCallback(
    async (id: string) => {
      const updated = await service.toggleDeviceStatus(id);
      setDevices(prev => prev.map(d => (d.id === id ? updated : d)));
      const evts = await service.getEvents();
      setEvents(evts);
    },
    [service]
  );

  const addNewSimulatedDevice = useCallback(async () => {
    const newDev = await service.addNewSimulatedDevice();
    setDevices(prev => [newDev, ...prev]);
    const evts = await service.getEvents();
    setEvents(evts);
    setSelectedDeviceId(newDev.id);
    setActivePage('device-details');
  }, [service]);

  const setPrototypeStatePreset = useCallback(
    async (preset: PrototypeStatePreset) => {
      setPrototypeState(preset);
      setError(null);

      switch (preset) {
        case 'normal':
          service.setSimulatedEmpty(false);
          service.setSimulatedError(false);
          await service.resetToDefault();
          setIsLoading(false);
          setIsScanning(false);
          await loadData();
          break;

        case 'scanning':
          service.setSimulatedEmpty(false);
          service.setSimulatedError(false);
          await loadData();
          setActivePage('scanner');
          startScan('quick');
          break;

        case 'scan_completed':
          service.setSimulatedEmpty(false);
          service.setSimulatedError(false);
          await loadData();
          setLastScanResult({
            scanId: 'scan-demo',
            type: 'quick',
            scannedAddresses: 24,
            totalAddresses: 24,
            devicesFound: 24,
            newDevices: 1,
            errors: 0,
            durationMs: 1420,
            timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
          });
          setActivePage('scanner');
          break;

        case 'new_device':
          service.setSimulatedEmpty(false);
          service.setSimulatedError(false);
          await loadData();
          // Find device with isNew or dev-10
          const newD = devices.find(d => d.isNew) || devices[9];
          if (newD) {
            setSelectedDeviceId(newD.id);
            setActivePage('device-details');
          }
          break;

        case 'device_offline':
          service.setSimulatedEmpty(false);
          service.setSimulatedError(false);
          await loadData();
          const offD = devices.find(d => d.status === 'offline') || devices[3];
          if (offD) {
            setSelectedDeviceId(offD.id);
            setActivePage('device-details');
          }
          break;

        case 'empty':
          service.setSimulatedError(false);
          service.setSimulatedEmpty(true);
          setDevices([]);
          setEvents([]);
          if (networkInfo) {
            setNetworkInfo({ ...networkInfo, activeAddresses: 0 });
          }
          setActivePage('dashboard');
          break;

        case 'loading':
          setIsLoading(true);
          setTimeout(() => {
            setIsLoading(false);
          }, 3000);
          break;

        case 'error':
          service.setSimulatedError(true);
          setError('NetWatch could not access the selected network interface. Interface unavailable or permissions restricted.');
          break;
      }
    },
    [service, loadData, startScan, devices, networkInfo]
  );

  const clearError = useCallback(() => {
    service.setSimulatedError(false);
    setError(null);
    loadData();
  }, [service, loadData]);

  const selectedDevice = devices.find(d => d.id === selectedDeviceId) || null;

  return (
    <NetworkContext.Provider
      value={{
        service,
        devices,
        networkInfo,
        events,
        selectedDevice,
        selectedDeviceId,
        activePage,
        isScanning,
        scanProgress,
        lastScanResult,
        isLoading,
        error,
        prototypeState,
        navigateTo,
        selectDevice,
        startScan,
        refresh,
        updateDevice,
        mergeDevices,
        pingDevice,
        wakeOnLan,
        scanDevicePorts,
        toggleDeviceStatus,
        addNewSimulatedDevice,
        setPrototypeStatePreset,
        clearError
      }}
    >
      {children}
    </NetworkContext.Provider>
  );
};

export const useNetwork = (): NetworkContextType => {
  const context = useContext(NetworkContext);
  if (!context) {
    throw new Error('useNetwork must be used within a NetworkProvider');
  }
  return context;
};
