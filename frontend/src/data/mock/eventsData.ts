import { NetworkEvent } from '../../types/events';

export const initialMockEvents: NetworkEvent[] = [
  {
    id: 'evt-01',
    timestamp: '23:14',
    type: 'new_device',
    title: 'New device discovered',
    deviceName: 'Xiaomi 14 Ultra',
    deviceId: 'dev-10',
    ip: '192.168.1.37',
    mac: '64:90:C1:28:FE:84',
    details: 'New host announced on subnet via DHCP'
  },
  {
    id: 'evt-02',
    timestamp: '22:51',
    type: 'online',
    title: 'Device came online',
    deviceName: 'Samsung Smart TV 65"',
    deviceId: 'dev-11',
    ip: '192.168.1.18',
    mac: 'AA:BB:CC:DD:EE:18',
    details: 'Resumed network connectivity (latency: 4ms)'
  },
  {
    id: 'evt-03',
    timestamp: '22:03',
    type: 'online',
    title: 'Device came online',
    deviceName: 'iPhone 16 Pro',
    deviceId: 'dev-06',
    ip: '192.168.1.24',
    mac: 'F0:18:98:C3:54:D2',
    details: 'Associated with SSID Home-5G'
  },
  {
    id: 'evt-04',
    timestamp: '21:44',
    type: 'network_change',
    title: 'Network Change',
    deviceName: 'Wi-Fi Interface',
    ip: '192.168.1.24',
    details: 'Interface re-negotiated to 802.11ax (866 Mbps link)'
  },
  {
    id: 'evt-05',
    timestamp: '19:32',
    type: 'offline',
    title: 'Device went offline',
    deviceName: 'ThinkPad T14',
    deviceId: 'dev-04',
    ip: '192.168.1.22',
    mac: 'AC:72:89:15:BE:33',
    details: 'No response to ARP / ICMP health check'
  },
  {
    id: 'evt-06',
    timestamp: '17:00',
    type: 'scan',
    title: 'Scheduled scan completed',
    details: '254 addresses surveyed · 24 devices discovered (0 errors)'
  },
  {
    id: 'evt-07',
    timestamp: '14:52',
    type: 'offline',
    title: 'Device went offline',
    deviceName: 'Galaxy Tab S9',
    deviceId: 'dev-09',
    ip: '192.168.1.30',
    details: 'Entered power saving sleep state'
  },
  {
    id: 'evt-08',
    timestamp: '12:15',
    type: 'service_change',
    title: 'Port detected open',
    deviceName: 'HomeLab MicroServer',
    deviceId: 'dev-05',
    ip: '192.168.1.50',
    details: 'Port 9000 (Portainer Docker) opened'
  }
];
