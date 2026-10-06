import { Device, DeviceEvent, DeviceService } from '../types/device';
import { NetworkInfo, ScanResult } from '../types/network';
import { NetworkEvent } from '../types/events';
import { AppSettings, DatabaseStats, MaintenanceResult, PruneResult } from '../types/settings';
import { NetworkService, ScanProgressCallback } from './NetworkService';

export interface EngineConnection {
  baseUrl: string;
  token: string;
}

/** Raw engine payloads: identical to the UI types except timestamps are RFC 3339. */
type WireDevice = Omit<Device, 'firstSeen' | 'lastSeen' | 'history'> & {
  firstSeen: string;
  lastSeen: string;
  history?: WireDeviceEvent[];
};
type WireDeviceEvent = DeviceEvent; // timestamp is an ISO string on the wire
type WireEvent = NetworkEvent;
type WireScanResult = ScanResult;

// ---- display formatting (the engine stays timezone-neutral) ----------------

const pad = (n: number) => String(n).padStart(2, '0');
const isSameDay = (a: Date, b: Date) => a.toDateString() === b.toDateString();

/** "23:14" today, "10-04 23:14" on earlier days. */
export function fmtClock(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  const hm = `${pad(d.getHours())}:${pad(d.getMinutes())}`;
  return isSameDay(d, new Date()) ? hm : `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${hm}`;
}

/** "2026-10-04 23:14" in local time. */
export function fmtDateTime(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function fmtRelative(iso: string, now = Date.now()): string {
  const t = new Date(iso).getTime();
  if (isNaN(t)) return iso;
  const s = Math.max(0, Math.round((now - t) / 1000));
  if (s < 150) return 'Now';
  const m = Math.round(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.round(m / 60);
  if (h < 24) return `${h}h ago`;
  return fmtDateTime(iso);
}

function toDevice(w: WireDevice): Device {
  return {
    ...w,
    firstSeen: fmtDateTime(w.firstSeen),
    lastSeen: w.status === 'online' ? 'Now' : fmtRelative(w.lastSeen),
    lastSeenAt: w.lastSeen,
    history: w.history?.map(toDeviceEvent),
  };
}
const toDeviceEvent = (e: WireDeviceEvent): DeviceEvent => ({ ...e, timestamp: fmtDateTime(e.timestamp) });
const toEvent = (e: WireEvent): NetworkEvent => ({ ...e, timestamp: fmtClock(e.timestamp) });
const toScanResult = (r: WireScanResult): ScanResult => ({ ...r, timestamp: fmtClock(r.timestamp) });

export class EngineError extends Error {
  constructor(message: string, public code = 'internal', public status = 0) {
    super(message);
  }
}

/** NetworkService backed by the local Go engine (127.0.0.1 + session token). */
export class HttpNetworkService implements NetworkService {
  readonly isSimulated = false;

  constructor(private conn: EngineConnection) {}

  private async request<T>(method: string, path: string, body?: unknown, signal?: AbortSignal): Promise<T> {
    let res: Response;
    try {
      res = await fetch(this.conn.baseUrl + path, {
        method,
        signal,
        headers: {
          Authorization: `Bearer ${this.conn.token}`,
          ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}),
        },
        body: body !== undefined ? JSON.stringify(body) : undefined,
      });
    } catch {
      throw new EngineError('NetWatch engine is not reachable. Restart the application.', 'unreachable');
    }
    if (res.status === 204) return undefined as T;
    const text = await res.text();
    let data: unknown;
    try { data = text ? JSON.parse(text) : undefined; } catch { data = undefined; }
    if (!res.ok) {
      const err = (data as { error?: { code?: string; message?: string } } | undefined)?.error;
      throw new EngineError(err?.message || `Engine returned HTTP ${res.status}`, err?.code, res.status);
    }
    return data as T;
  }

  async getNetworkInfo(): Promise<NetworkInfo> {
    return this.request<NetworkInfo>('GET', '/v1/network');
  }

  async getDevices(): Promise<Device[]> {
    return (await this.request<WireDevice[]>('GET', '/v1/devices')).map(toDevice);
  }

  async getDevice(id: string): Promise<Device | undefined> {
    try {
      return toDevice(await this.request<WireDevice>('GET', `/v1/devices/${encodeURIComponent(id)}`));
    } catch (e) {
      if (e instanceof EngineError && e.status === 404) return undefined;
      throw e;
    }
  }

  async getEvents(): Promise<NetworkEvent[]> {
    return (await this.request<WireEvent[]>('GET', '/v1/events')).map(toEvent);
  }

  async getDeviceHistory(id: string): Promise<DeviceEvent[]> {
    return (await this.request<WireDeviceEvent[]>('GET', `/v1/devices/${encodeURIComponent(id)}/history`)).map(toDeviceEvent);
  }

  async updateDevice(id: string, updates: Partial<Device>): Promise<Device> {
    const patch: Record<string, unknown> = {};
    // An absent value means "clear it", so send empty strings explicitly.
    if ('customAlias' in updates) patch.customAlias = updates.customAlias ?? '';
    if ('notes' in updates) patch.notes = updates.notes ?? '';
    if ('isNew' in updates) patch.isNew = !!updates.isNew;
    if ('trustStatus' in updates) patch.trustStatus = updates.trustStatus;
    return toDevice(await this.request<WireDevice>('PATCH', `/v1/devices/${encodeURIComponent(id)}`, patch));
  }

  async mergeDevices(targetId: string, sourceId: string): Promise<void> {
    await this.request('POST', `/v1/devices/${encodeURIComponent(targetId)}/merge`, { sourceId });
  }

  async pingDevice(ip: string): Promise<{ success: boolean; latencyMs: number }> {
    return this.request('POST', '/v1/ping', { ip });
  }

  async wakeOnLan(mac: string): Promise<{ success: boolean; message: string }> {
    return this.request('POST', '/v1/wol', { mac });
  }

  async scanDevicePorts(id: string): Promise<DeviceService[]> {
    return (await this.request<DeviceService[] | null>('POST', `/v1/devices/${encodeURIComponent(id)}/ports`)) ?? [];
  }

  async scanNetwork(type: 'quick' | 'full' = 'quick', onProgress?: ScanProgressCallback): Promise<ScanResult> {
    const { scanId } = await this.request<{ scanId: string }>('POST', '/v1/scans', { type });
    const res = await fetch(`${this.conn.baseUrl}/v1/scans/${encodeURIComponent(scanId)}/stream`, {
      headers: { Authorization: `Bearer ${this.conn.token}`, Accept: 'text/event-stream' },
    }).catch(() => { throw new EngineError('NetWatch engine is not reachable. Restart the application.', 'unreachable'); });
    if (!res.ok || !res.body) throw new EngineError(`Scan stream failed (HTTP ${res.status})`, 'internal', res.status);

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buf = '';
    for (;;) {
      const { value, done } = await reader.read();
      if (done) break;
      buf += decoder.decode(value, { stream: true });
      let sep: number;
      while ((sep = buf.indexOf('\n\n')) >= 0) {
        const frame = buf.slice(0, sep);
        buf = buf.slice(sep + 2);
        let event = 'message';
        let data = '';
        for (const line of frame.split('\n')) {
          if (line.startsWith('event:')) event = line.slice(6).trim();
          else if (line.startsWith('data:')) data += line.slice(5).trim();
        }
        if (!data) continue; // keep-alive comment
        const payload = JSON.parse(data);
        if (event === 'progress') {
          onProgress?.(payload.progress, payload.scanned, payload.total, payload.found);
        } else if (event === 'done') {
          reader.cancel().catch(() => {});
          return toScanResult(payload as WireScanResult);
        } else if (event === 'scan-error') {
          throw new EngineError(payload.message || 'Network scan failed', 'scan_failed');
        }
      }
    }
    throw new EngineError('Scan stream ended unexpectedly', 'internal');
  }

  async clearHistory(): Promise<void> {
    await this.request('DELETE', '/v1/data');
  }

  async openDataFolder(): Promise<void> {
    await this.request('POST', '/v1/data/open');
  }

  async getSettings(): Promise<AppSettings> {
    return this.request<AppSettings>('GET', '/v1/settings');
  }

  async updateSettings(settings: Partial<AppSettings>): Promise<AppSettings> {
    return this.request<AppSettings>('PUT', '/v1/settings', settings);
  }

  async getDatabaseStats(): Promise<DatabaseStats> {
    return this.request<DatabaseStats>('GET', '/v1/data/stats');
  }

  async vacuumDatabase(): Promise<MaintenanceResult> {
    return this.request<MaintenanceResult>('POST', '/v1/data/vacuum');
  }

  async integrityCheck(): Promise<MaintenanceResult> {
    return this.request<MaintenanceResult>('POST', '/v1/data/integrity');
  }

  async pruneEvents(olderThanDays = 30): Promise<PruneResult> {
    return this.request<PruneResult>('POST', '/v1/data/prune', { olderThanDays });
  }

  // Prototype-only helpers: meaningless against a real network.
  setSimulatedEmpty(): void {}
  setSimulatedError(): void {}
  async toggleDeviceStatus(): Promise<Device> { throw new EngineError('Not available on a real network', 'unsupported'); }
  async addNewSimulatedDevice(): Promise<Device> { throw new EngineError('Not available on a real network', 'unsupported'); }
  async resetToDefault(): Promise<void> {}
}
