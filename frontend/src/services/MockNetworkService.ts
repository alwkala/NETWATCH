import { Device, DeviceEvent, DeviceService } from '../types/device';
import { NetworkInfo, ScanResult } from '../types/network';
import { NetworkEvent } from '../types/events';
import { initialMockDevices } from '../data/mock/devicesData';
import { initialMockNetworkInfo } from '../data/mock/networkData';
import { initialMockEvents } from '../data/mock/eventsData';
import { NetworkService, ScanProgressCallback } from './NetworkService';

export class MockNetworkService implements NetworkService {
  public readonly isSimulated = true;
  private devices: Device[];
  private networkInfo: NetworkInfo;
  private events: NetworkEvent[];
  private isSimulatedEmpty = false;
  private isSimulatedError = false;

  constructor() {
    this.devices = JSON.parse(JSON.stringify(initialMockDevices));
    this.networkInfo = JSON.parse(JSON.stringify(initialMockNetworkInfo));
    this.events = JSON.parse(JSON.stringify(initialMockEvents));
  }

  public setSimulatedEmpty(empty: boolean): void {
    this.isSimulatedEmpty = empty;
  }

  public setSimulatedError(error: boolean): void {
    this.isSimulatedError = error;
  }

  public async getNetworkInfo(): Promise<NetworkInfo> {
    await this.delay(100);
    if (this.isSimulatedError) {
      throw new Error('NetWatch could not access the selected network interface. Interface unavailable or permissions restricted.');
    }
    if (this.isSimulatedEmpty) {
      return {
        ...this.networkInfo,
        activeAddresses: 0,
        status: 'Connected'
      };
    }
    const onlineCount = this.devices.filter(d => d.status === 'online').length;
    return {
      ...this.networkInfo,
      activeAddresses: onlineCount
    };
  }

  public async getDevices(): Promise<Device[]> {
    await this.delay(150);
    if (this.isSimulatedError) {
      throw new Error('Unable to enumerate ARP/NDP tables on interface Wi-Fi.');
    }
    if (this.isSimulatedEmpty) {
      return [];
    }
    return [...this.devices];
  }

  public async getDevice(id: string): Promise<Device | undefined> {
    await this.delay(80);
    return this.devices.find(d => d.id === id);
  }

  public async scanNetwork(type: 'quick' | 'full' = 'quick', onProgress?: ScanProgressCallback): Promise<ScanResult> {
    if (this.isSimulatedError) {
      await this.delay(300);
      throw new Error('Network interface returned socket access failure during ARP sweep.');
    }

    const totalSteps = type === 'quick' ? 12 : 20;
    const totalAddresses = 24;

    for (let i = 1; i <= totalSteps; i++) {
      await this.delay(type === 'quick' ? 120 : 180);
      const progress = Math.round((i / totalSteps) * 100);
      const scanned = Math.min(24, Math.round((i / totalSteps) * totalAddresses));
      const found = Math.min(this.devices.length, Math.max(1, Math.round(scanned * 0.95)));
      if (onProgress) {
        onProgress(progress, scanned, totalAddresses, found);
      }
    }

    // If was empty, populate upon scan!
    if (this.isSimulatedEmpty) {
      this.isSimulatedEmpty = false;
    }

    const onlineDevices = this.devices.filter(d => d.status === 'online').length;
    const newDevices = this.devices.filter(d => d.isNew).length;

    const result: ScanResult = {
      scanId: `scan-${Date.now()}`,
      type,
      scannedAddresses: 24,
      totalAddresses: 24,
      devicesFound: this.devices.length,
      newDevices: newDevices || 1,
      errors: 0,
      durationMs: type === 'quick' ? 1440 : 3600,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    };

    // Add scan event to events
    const scanEvent: NetworkEvent = {
      id: `evt-${Date.now()}`,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      type: 'scan',
      title: `${type === 'quick' ? 'Quick' : 'Full'} Discovery Scan Completed`,
      details: `${result.scannedAddresses} addresses scanned · ${result.devicesFound} devices discovered (0 errors)`
    };
    this.events.unshift(scanEvent);

    return result;
  }

  public async getEvents(): Promise<NetworkEvent[]> {
    await this.delay(80);
    if (this.isSimulatedEmpty) return [];
    return [...this.events];
  }

  public async getDeviceHistory(id: string): Promise<DeviceEvent[]> {
    await this.delay(60);
    const dev = this.devices.find(d => d.id === id);
    return dev?.history || [];
  }

  public async updateDevice(id: string, updates: Partial<Device>): Promise<Device> {
    await this.delay(100);
    const index = this.devices.findIndex(d => d.id === id);
    if (index === -1) {
      throw new Error(`Device ${id} not found`);
    }
    this.devices[index] = { ...this.devices[index], ...updates };
    return this.devices[index];
  }

  public async pingDevice(ip: string): Promise<{ success: boolean; latencyMs: number }> {
    await this.delay(280);
    const dev = this.devices.find(d => d.ip === ip);
    if (!dev || dev.status === 'offline') {
      return { success: false, latencyMs: 0 };
    }
    // Realistic jitter between 1 and 25ms
    const base = dev.latencyMs || 5;
    const jitter = Math.floor(Math.random() * 4) - 2;
    const simulatedLatency = Math.max(1, base + jitter);
    return { success: true, latencyMs: simulatedLatency };
  }

  public async wakeOnLan(mac: string): Promise<{ success: boolean; message: string }> {
    await this.delay(400);
    return {
      success: true,
      message: `Magic packet (102 bytes) broadcasted to FF:FF:FF:FF:FF:FF for MAC ${mac} on port 9.`
    };
  }

  public async scanDevicePorts(id: string): Promise<DeviceService[]> {
    await this.delay(700);
    const dev = this.devices.find(d => d.id === id);
    if (!dev) return [];
    if (dev.status === 'offline') return [];
    if (!dev.services || dev.services.length === 0) {
      // Return a basic HTTP/HTTPS default or empty
      return [
        { port: 80, protocol: 'TCP', service: 'HTTP', status: 'Closed' },
        { port: 443, protocol: 'TCP', service: 'HTTPS', status: 'Closed' }
      ];
    }
    return [...dev.services];
  }

  public async toggleDeviceStatus(id: string): Promise<Device> {
    const dev = this.devices.find(d => d.id === id);
    if (!dev) throw new Error('Device not found');
    const newStatus = dev.status === 'online' ? 'offline' : 'online';
    dev.status = newStatus;
    dev.lastSeen = newStatus === 'online' ? 'Now' : '1m ago';
    if (newStatus === 'online') {
      dev.latencyMs = 5;
    } else {
      dev.latencyMs = undefined;
    }

    const timeStr = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    const event: NetworkEvent = {
      id: `evt-${Date.now()}`,
      timestamp: timeStr,
      type: newStatus === 'online' ? 'online' : 'offline',
      title: newStatus === 'online' ? 'Device came online' : 'Device went offline',
      deviceName: dev.name,
      deviceId: dev.id,
      ip: dev.ip,
      mac: dev.mac,
      details: newStatus === 'online' ? 'Responded to ICMP echo' : 'No response to ARP keepalive'
    };
    this.events.unshift(event);

    if (!dev.history) dev.history = [];
    dev.history.unshift({
      id: `h-${Date.now()}`,
      timestamp: `Today ${timeStr}`,
      type: newStatus === 'online' ? 'online' : 'offline',
      description: `Device transitioned to ${newStatus}`
    });

    return { ...dev };
  }

  public async addNewSimulatedDevice(): Promise<Device> {
    const nextOctet = 120 + Math.floor(Math.random() * 50);
    const newDev: Device = {
      id: `dev-${Date.now()}`,
      name: `Raspberry Pi 5 (${nextOctet})`,
      hostname: `rpi5-node-${nextOctet}.local`,
      ip: `192.168.1.${nextOctet}`,
      mac: `B8:27:EB:${Math.floor(Math.random()*89+10)}:${Math.floor(Math.random()*89+10)}:${Math.floor(Math.random()*89+10)}`,
      vendor: 'Raspberry Pi Trading Ltd',
      type: 'Computer',
      status: 'online',
      isNew: true,
      firstSeen: 'Just now',
      lastSeen: 'Now',
      latencyMs: 3,
      os: 'Raspberry Pi OS 64-bit',
      services: [
        { port: 22, protocol: 'TCP', service: 'SSH', status: 'Open' },
        { port: 80, protocol: 'TCP', service: 'HTTP', status: 'Open' }
      ],
      history: [
        { id: `h-${Date.now()}`, timestamp: 'Just now', type: 'discovered', description: 'Discovered via DHCP probe' }
      ]
    };

    this.devices.unshift(newDev);
    const event: NetworkEvent = {
      id: `evt-${Date.now()}`,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      type: 'new_device',
      title: 'New device discovered',
      deviceName: newDev.name,
      deviceId: newDev.id,
      ip: newDev.ip,
      mac: newDev.mac,
      details: 'New hardware MAC registered on subnet'
    };
    this.events.unshift(event);
    return newDev;
  }

  public async resetToDefault(): Promise<void> {
    this.devices = JSON.parse(JSON.stringify(initialMockDevices));
    this.networkInfo = JSON.parse(JSON.stringify(initialMockNetworkInfo));
    this.events = JSON.parse(JSON.stringify(initialMockEvents));
    this.isSimulatedEmpty = false;
    this.isSimulatedError = false;
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

export const defaultNetworkService = new MockNetworkService();
