export type EventType =
  | 'new_device'
  | 'online'
  | 'offline'
  | 'network_change'
  | 'scan'
  | 'service_change';

export interface NetworkEvent {
  id: string;
  timestamp: string;
  type: EventType;
  title: string;
  deviceName?: string;
  deviceId?: string;
  ip?: string;
  mac?: string;
  details?: string;
}
