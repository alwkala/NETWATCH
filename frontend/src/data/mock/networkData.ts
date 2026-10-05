import { NetworkInfo } from '../../types/network';

export const initialMockNetworkInfo: NetworkInfo = {
  networkName: 'Home Network',
  ssid: 'Home-5G',
  interfaceName: 'Wi-Fi',
  interfaceType: 'Wi-Fi',
  gateway: '192.168.1.1',
  subnet: '192.168.1.0/24',
  dns: ['192.168.1.1', '1.1.1.1'],
  localIp: '192.168.1.24',
  broadcast: '192.168.1.255',
  netmask: '255.255.255.0',
  totalAddresses: 254,
  activeAddresses: 24,
  status: 'Connected',
  pingStats: {
    currentMs: 4,
    avgMs: 7,
    peakMs: 21,
    packetLossPercent: 0,
    history: [4, 5, 7, 4, 6, 8, 21, 6, 5, 4, 4, 7, 5, 4, 6, 5, 4, 8, 6, 4]
  },
  interfaces: [
    {
      id: 'if-wifi',
      name: 'Wi-Fi',
      type: 'Wi-Fi',
      adapterName: 'Intel(R) Wi-Fi 6E AX211 160MHz',
      ip: '192.168.1.24',
      subnet: '192.168.1.0/24',
      gateway: '192.168.1.1',
      mac: 'F0:18:98:C3:54:D2',
      status: 'Connected',
      speedMbps: 866,
      isDefault: true
    },
    {
      id: 'if-eth',
      name: 'Ethernet',
      type: 'Ethernet',
      adapterName: 'Realtek PCIe GbE Family Controller',
      ip: '0.0.0.0',
      subnet: '—',
      gateway: '—',
      mac: '70:B5:E8:2C:91:05',
      status: 'Disconnected',
      speedMbps: 0,
      isDefault: false
    },
    {
      id: 'if-vpn',
      name: 'Tailscale Tunnel',
      type: 'VPN',
      adapterName: 'Tailscale Tunnel Adapter',
      ip: '100.84.19.42',
      subnet: '100.64.0.0/10',
      gateway: '—',
      mac: '00:00:00:00:00:00',
      status: 'Disconnected',
      speedMbps: 100,
      isDefault: false
    }
  ]
};
