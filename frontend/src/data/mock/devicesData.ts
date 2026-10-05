import { Device } from '../../types/device';

export const initialMockDevices: Device[] = [
  // 1 Router
  {
    id: 'dev-01',
    name: 'TP-Link Archer AX73',
    hostname: 'router.local',
    ip: '192.168.1.1',
    mac: 'E8:48:B8:31:7A:01',
    vendor: 'TP-Link Technologies',
    type: 'Router',
    status: 'online',
    firstSeen: '2026-08-10 08:00',
    lastSeen: 'Now',
    latencyMs: 1,
    os: 'Embedded Linux (TP-Link Firmware 1.3.2)',
    notes: 'Default Gateway & DHCP Server',
    services: [
      { port: 53, protocol: 'UDP', service: 'DNS (Dnsmasq)', status: 'Open' },
      { port: 80, protocol: 'TCP', service: 'HTTP (Web Admin)', status: 'Open' },
      { port: 443, protocol: 'TCP', service: 'HTTPS (Web Admin SSL)', status: 'Open' },
      { port: 1900, protocol: 'UDP', service: 'UPnP / SSDP', status: 'Open' }
    ],
    history: [
      { id: 'h1', timestamp: '2026-10-03 17:50', type: 'online', description: 'Health check OK (1ms)' },
      { id: 'h2', timestamp: '2026-08-10 08:00', type: 'discovered', description: 'Router identified as primary gateway' }
    ]
  },

  // 4 Computers
  {
    id: 'dev-02',
    name: 'MacBook Pro 16"',
    hostname: 'adrian-mbp.local',
    ip: '192.168.1.12',
    mac: '3C:06:30:4A:21:8F',
    vendor: 'Apple, Inc.',
    type: 'Computer',
    status: 'online',
    firstSeen: '2026-09-01 09:12',
    lastSeen: 'Now',
    latencyMs: 3,
    os: 'macOS 15.2 (Sequoia)',
    services: [
      { port: 22, protocol: 'TCP', service: 'SSH (OpenSSH 9.8)', status: 'Open' },
      { port: 5353, protocol: 'UDP', service: 'mDNS (Bonjour)', status: 'Open' },
      { port: 5000, protocol: 'TCP', service: 'AirPlay Receiver', status: 'Open' }
    ],
    history: [
      { id: 'h3', timestamp: '2026-10-03 16:30', type: 'online', description: 'Device came online' },
      { id: 'h4', timestamp: '2026-10-03 12:15', type: 'offline', description: 'Device went offline (sleep)' }
    ]
  },
  {
    id: 'dev-03',
    name: 'Dell XPS 15 Desktop/Laptop',
    hostname: 'workstation-win11',
    ip: '192.168.1.15',
    mac: '70:B5:E8:2C:91:04',
    vendor: 'Dell Inc.',
    type: 'Computer',
    status: 'online',
    firstSeen: '2026-09-05 11:40',
    lastSeen: 'Now',
    latencyMs: 4,
    os: 'Windows 11 Pro 24H2',
    notes: 'Primary dev workstation',
    services: [
      { port: 135, protocol: 'TCP', service: 'MSRPC', status: 'Open' },
      { port: 445, protocol: 'TCP', service: 'SMB (File Sharing)', status: 'Open' },
      { port: 3389, protocol: 'TCP', service: 'RDP (Remote Desktop)', status: 'Open' }
    ],
    history: [
      { id: 'h5', timestamp: '2026-10-03 08:30', type: 'online', description: 'Connected to Wi-Fi' }
    ]
  },
  {
    id: 'dev-04',
    name: 'ThinkPad T14',
    hostname: 'sarah-thinkpad',
    ip: '192.168.1.22',
    mac: 'AC:72:89:15:BE:33',
    vendor: 'Lenovo',
    type: 'Computer',
    status: 'offline',
    firstSeen: '2026-09-12 14:00',
    lastSeen: '42m ago',
    latencyMs: undefined,
    os: 'Ubuntu 24.04 LTS',
    services: [
      { port: 22, protocol: 'TCP', service: 'SSH', status: 'Closed' }
    ],
    history: [
      { id: 'h6', timestamp: '2026-10-03 17:14', type: 'offline', description: 'Device went offline' },
      { id: 'h7', timestamp: '2026-10-03 09:02', type: 'online', description: 'Device came online' }
    ]
  },
  {
    id: 'dev-05',
    name: 'HomeLab MicroServer',
    hostname: 'srv-vault.local',
    ip: '192.168.1.50',
    mac: '00:1E:67:84:9D:1A',
    vendor: 'Intel Corporate',
    type: 'Server',
    status: 'online',
    firstSeen: '2026-08-15 10:00',
    lastSeen: 'Now',
    latencyMs: 2,
    os: 'Debian 12 Bookworm',
    notes: 'Local backup & media server',
    services: [
      { port: 22, protocol: 'TCP', service: 'SSH', status: 'Open' },
      { port: 80, protocol: 'TCP', service: 'HTTP (Nginx reverse proxy)', status: 'Open' },
      { port: 443, protocol: 'TCP', service: 'HTTPS', status: 'Open' },
      { port: 8080, protocol: 'TCP', service: 'Dashboard / API', status: 'Open' },
      { port: 9000, protocol: 'TCP', service: 'Portainer Docker', status: 'Open' }
    ],
    history: [
      { id: 'h8', timestamp: '2026-10-03 00:00', type: 'online', description: 'Daily uptime check passed' }
    ]
  },

  // 5 Phones / Tablets
  {
    id: 'dev-06',
    name: 'iPhone 16 Pro',
    hostname: 'Alex-iPhone',
    ip: '192.168.1.24',
    mac: 'F0:18:98:C3:54:D2',
    vendor: 'Apple, Inc.',
    type: 'Phone',
    status: 'online',
    firstSeen: '2026-09-20 18:22',
    lastSeen: 'Now',
    latencyMs: 8,
    os: 'iOS 19.1',
    services: [
      { port: 5353, protocol: 'UDP', service: 'mDNS (Bonjour)', status: 'Open' }
    ],
    history: [
      { id: 'h9', timestamp: '2026-10-03 15:40', type: 'online', description: 'Connected to Home-5G' }
    ]
  },
  {
    id: 'dev-07',
    name: 'iPad Pro 11"',
    hostname: 'Studio-iPad',
    ip: '192.168.1.26',
    mac: 'BC:D1:D3:8E:77:50',
    vendor: 'Apple, Inc.',
    type: 'Tablet',
    status: 'online',
    firstSeen: '2026-09-18 16:04',
    lastSeen: 'Now',
    latencyMs: 11,
    os: 'iPadOS 19.0',
    services: [
      { port: 5353, protocol: 'UDP', service: 'mDNS (Bonjour)', status: 'Open' }
    ],
    history: [
      { id: 'h10', timestamp: '2026-10-03 14:10', type: 'online', description: 'Device came online' }
    ]
  },
  {
    id: 'dev-08',
    name: 'Pixel 9 Pro',
    hostname: 'pixel-9-adrian',
    ip: '192.168.1.28',
    mac: '58:24:29:4B:91:E7',
    vendor: 'Google, LLC',
    type: 'Phone',
    status: 'online',
    firstSeen: '2026-09-02 12:45',
    lastSeen: 'Now',
    latencyMs: 9,
    os: 'Android 15',
    services: [
      { port: 5353, protocol: 'UDP', service: 'mDNS', status: 'Open' }
    ],
    history: [
      { id: 'h11', timestamp: '2026-10-03 17:35', type: 'online', description: 'Ping response 9ms' }
    ]
  },
  {
    id: 'dev-09',
    name: 'Galaxy Tab S9',
    hostname: 'galaxy-tab-livingroom',
    ip: '192.168.1.30',
    mac: '24:4B:FE:09:A4:71',
    vendor: 'Samsung Electronics',
    type: 'Tablet',
    status: 'offline',
    firstSeen: '2026-09-22 19:10',
    lastSeen: '3h ago',
    latencyMs: undefined,
    os: 'One UI 6.1 (Android 14)',
    services: [],
    history: [
      { id: 'h12', timestamp: '2026-10-03 14:52', type: 'offline', description: 'Device went offline (Wi-Fi sleep)' }
    ]
  },
  {
    id: 'dev-10',
    name: 'Xiaomi 14 Ultra (New Device)',
    hostname: 'xiaomi-device-37',
    ip: '192.168.1.37',
    mac: '64:90:C1:28:FE:84',
    vendor: 'Xiaomi Communications',
    type: 'Phone',
    status: 'online',
    isNew: true,
    firstSeen: '2026-10-03 17:54',
    lastSeen: '2m ago',
    latencyMs: 14,
    os: 'Xiaomi HyperOS 2.0',
    notes: 'Recently connected guest device',
    services: [
      { port: 5353, protocol: 'UDP', service: 'mDNS', status: 'Open' }
    ],
    history: [
      { id: 'h13', timestamp: '2026-10-03 17:54', type: 'discovered', description: 'New device discovered via ARP broadcast' },
      { id: 'h14', timestamp: '2026-10-03 17:54', type: 'online', description: 'Assigned DHCP lease 192.168.1.37' }
    ]
  },

  // 2 TVs
  {
    id: 'dev-11',
    name: 'Samsung Smart TV 65"',
    hostname: 'livingroom-tv',
    ip: '192.168.1.18',
    mac: 'AA:BB:CC:DD:EE:18',
    vendor: 'Samsung Electronics',
    type: 'TV',
    status: 'online',
    firstSeen: '2026-09-14 10:21',
    lastSeen: '2m ago',
    latencyMs: 4,
    os: 'Tizen OS 8.0',
    services: [
      { port: 80, protocol: 'TCP', service: 'HTTP (DIAL Protocol)', status: 'Open' },
      { port: 443, protocol: 'TCP', service: 'HTTPS (Web API)', status: 'Open' },
      { port: 8008, protocol: 'TCP', service: 'HTTP (SmartView / Cast)', status: 'Open' },
      { port: 8443, protocol: 'TCP', service: 'HTTPS (SmartView Secure)', status: 'Open' },
      { port: 9197, protocol: 'TCP', service: 'Samsung Remote Control', status: 'Open' }
    ],
    history: [
      { id: 'h15', timestamp: '2026-10-03 17:38', type: 'online', description: 'Device came online' },
      { id: 'h16', timestamp: '2026-10-02 23:40', type: 'offline', description: 'Device went offline' },
      { id: 'h17', timestamp: '2026-09-14 10:21', type: 'discovered', description: 'First discovered on subnet' }
    ]
  },
  {
    id: 'dev-12',
    name: 'Sony Bravia 4K Bedroom',
    hostname: 'sony-bravia-bed',
    ip: '192.168.1.19',
    mac: '00:24:8D:67:3B:55',
    vendor: 'Sony Corporation',
    type: 'TV',
    status: 'online',
    firstSeen: '2026-09-15 19:30',
    lastSeen: 'Now',
    latencyMs: 7,
    os: 'Google TV / Android TV 14',
    services: [
      { port: 8008, protocol: 'TCP', service: 'Google Cast', status: 'Open' },
      { port: 8443, protocol: 'TCP', service: 'Cast Secure', status: 'Open' },
      { port: 9000, protocol: 'TCP', service: 'Sony IRCC Protocol', status: 'Open' }
    ],
    history: [
      { id: 'h18', timestamp: '2026-10-03 16:00', type: 'online', description: 'Device came online' }
    ]
  },

  // 2 Printers
  {
    id: 'dev-13',
    name: 'HP LaserJet Pro M404dw',
    hostname: 'hpoffice-m404.local',
    ip: '192.168.1.40',
    mac: 'B4:99:BA:78:23:41',
    vendor: 'HP Inc.',
    type: 'Printer',
    status: 'online',
    firstSeen: '2026-08-20 09:00',
    lastSeen: 'Now',
    latencyMs: 5,
    os: 'HP FutureSmart Firmware',
    services: [
      { port: 80, protocol: 'TCP', service: 'HTTP (Embedded Web Server)', status: 'Open' },
      { port: 443, protocol: 'TCP', service: 'HTTPS (EWS)', status: 'Open' },
      { port: 631, protocol: 'TCP', service: 'IPP (Internet Printing Protocol)', status: 'Open' },
      { port: 9100, protocol: 'TCP', service: 'RAW JetDirect', status: 'Open' }
    ],
    history: [
      { id: 'h19', timestamp: '2026-10-03 08:00', type: 'online', description: 'Printer ready / paper OK' }
    ]
  },
  {
    id: 'dev-14',
    name: 'Brother HL-L2350DW',
    hostname: 'brother-lab.local',
    ip: '192.168.1.41',
    mac: '30:05:5C:8B:11:F2',
    vendor: 'Brother Industries, Ltd.',
    type: 'Printer',
    status: 'online',
    firstSeen: '2026-08-25 15:20',
    lastSeen: '5m ago',
    latencyMs: 6,
    os: 'Brother Embedded Print Server',
    services: [
      { port: 80, protocol: 'TCP', service: 'HTTP (Maintenance page)', status: 'Open' },
      { port: 631, protocol: 'TCP', service: 'IPP', status: 'Open' },
      { port: 9100, protocol: 'TCP', service: 'JetDirect', status: 'Open' }
    ],
    history: [
      { id: 'h20', timestamp: '2026-10-03 17:51', type: 'online', description: 'Status poll OK' }
    ]
  },

  // 3 IoT Devices
  {
    id: 'dev-15',
    name: 'Philips Hue Bridge v2',
    hostname: 'philips-hue-bridge',
    ip: '192.168.1.60',
    mac: '00:17:88:5A:D9:31',
    vendor: 'Signify Netherlands B.V.',
    type: 'IoT',
    status: 'online',
    firstSeen: '2026-08-11 11:00',
    lastSeen: 'Now',
    latencyMs: 3,
    os: 'Hue Embedded Linux',
    services: [
      { port: 80, protocol: 'TCP', service: 'HTTP (Hue API v2)', status: 'Open' },
      { port: 443, protocol: 'TCP', service: 'HTTPS (Hue API)', status: 'Open' },
      { port: 2100, protocol: 'UDP', service: 'Hue Entertainment Stream', status: 'Open' }
    ],
    history: [
      { id: 'h21', timestamp: '2026-10-03 17:00', type: 'online', description: 'Heartbeat OK' }
    ]
  },
  {
    id: 'dev-16',
    name: 'Ecobee Smart Thermostat',
    hostname: 'ecobee-hallway',
    ip: '192.168.1.62',
    mac: '44:61:32:89:33:0A',
    vendor: 'Ecobee Inc.',
    type: 'IoT',
    status: 'online',
    firstSeen: '2026-08-12 16:30',
    lastSeen: 'Now',
    latencyMs: 12,
    os: 'FreeRTOS / Embedded',
    services: [
      { port: 5353, protocol: 'UDP', service: 'mDNS (HomeKit)', status: 'Open' }
    ],
    history: [
      { id: 'h22', timestamp: '2026-10-03 17:30', type: 'online', description: 'Telemetry poll OK' }
    ]
  },
  {
    id: 'dev-17',
    name: 'Sonos One Speaker',
    hostname: 'sonos-kitchen.local',
    ip: '192.168.1.65',
    mac: '48:A6:B8:21:40:99',
    vendor: 'Sonos, Inc.',
    type: 'IoT',
    status: 'online',
    firstSeen: '2026-08-18 19:15',
    lastSeen: 'Now',
    latencyMs: 6,
    os: 'Sonos OS (Linux)',
    services: [
      { port: 1400, protocol: 'TCP', service: 'UPnP Sonos Control', status: 'Open' },
      { port: 1443, protocol: 'TCP', service: 'HTTPS Sonos Secure API', status: 'Open' }
    ],
    history: [
      { id: 'h23', timestamp: '2026-10-03 16:45', type: 'online', description: 'AirPlay connection active' }
    ]
  },

  // 2 Network Devices
  {
    id: 'dev-18',
    name: 'Netgear GS308T Managed Switch',
    hostname: 'switch-rack-01',
    ip: '192.168.1.2',
    mac: '28:80:88:AC:32:11',
    vendor: 'Netgear Inc.',
    type: 'Network Device',
    status: 'online',
    firstSeen: '2026-08-10 08:05',
    lastSeen: 'Now',
    latencyMs: 1,
    os: 'Netgear Smart Managed OS',
    services: [
      { port: 80, protocol: 'TCP', service: 'HTTP (Switch GUI)', status: 'Open' },
      { port: 161, protocol: 'UDP', service: 'SNMP', status: 'Open' }
    ],
    history: [
      { id: 'h24', timestamp: '2026-10-03 17:00', type: 'online', description: 'Port status 8/8 Gigabit Link Up' }
    ]
  },
  {
    id: 'dev-19',
    name: 'UniFi U6 Pro Access Point',
    hostname: 'u6-pro-hallway',
    ip: '192.168.1.3',
    mac: '74:83:C2:59:71:EE',
    vendor: 'Ubiquiti Inc.',
    type: 'Network Device',
    status: 'online',
    firstSeen: '2026-08-10 08:10',
    lastSeen: 'Now',
    latencyMs: 2,
    os: 'UniFi OS v3.2',
    services: [
      { port: 22, protocol: 'TCP', service: 'SSH', status: 'Open' },
      { port: 8080, protocol: 'TCP', service: 'UniFi Controller Inform', status: 'Open' }
    ],
    history: [
      { id: 'h25', timestamp: '2026-10-03 17:50', type: 'online', description: 'Channel 36 (5GHz) / Channel 6 (2.4GHz) nominal' }
    ]
  },

  // 1 Camera
  {
    id: 'dev-20',
    name: 'Reolink E1 Pro Security Cam',
    hostname: 'cam-front-door',
    ip: '192.168.1.70',
    mac: 'EC:71:DB:44:88:9C',
    vendor: 'Reolink Innovation',
    type: 'Camera',
    status: 'online',
    firstSeen: '2026-08-22 13:40',
    lastSeen: 'Now',
    latencyMs: 5,
    os: 'Embedded Linux RTSP Streamer',
    services: [
      { port: 80, protocol: 'TCP', service: 'HTTP Web View', status: 'Open' },
      { port: 554, protocol: 'TCP', service: 'RTSP Video Stream', status: 'Open' },
      { port: 8000, protocol: 'TCP', service: 'ONVIF Media Service', status: 'Open' }
    ],
    history: [
      { id: 'h26', timestamp: '2026-10-03 17:48', type: 'online', description: 'RTSP video stream active' }
    ]
  },

  // 4 Unknown / Other devices
  {
    id: 'dev-21',
    name: 'Espressif NodeMCU ESP32',
    hostname: 'esp32-sensor-garage',
    ip: '192.168.1.88',
    mac: '24:0A:C4:F3:19:D4',
    vendor: 'Espressif Inc.',
    type: 'IoT',
    status: 'online',
    firstSeen: '2026-09-08 17:10',
    lastSeen: 'Now',
    latencyMs: 8,
    os: 'ESP-IDF / FreeRTOS',
    services: [
      { port: 80, protocol: 'TCP', service: 'HTTP (Sensor Telemetry)', status: 'Open' }
    ],
    history: [
      { id: 'h27', timestamp: '2026-10-03 17:52', type: 'online', description: 'Temperature reading transmitted' }
    ]
  },
  {
    id: 'dev-22',
    name: 'Unknown Host (.94)',
    hostname: 'unknown-00-1B-44',
    ip: '192.168.1.94',
    mac: '00:1B:44:11:92:B3',
    vendor: 'SanDisk Corporation',
    type: 'Unknown',
    status: 'online',
    firstSeen: '2026-09-28 21:05',
    lastSeen: '8m ago',
    latencyMs: 18,
    services: [],
    history: [
      { id: 'h28', timestamp: '2026-10-03 17:48', type: 'online', description: 'Ping response received' }
    ]
  },
  {
    id: 'dev-23',
    name: 'Synology DiskStation DS920+',
    hostname: 'nas-storage.local',
    ip: '192.168.1.100',
    mac: '00:11:32:8A:F1:C9',
    vendor: 'Synology Inc.',
    type: 'Server',
    status: 'online',
    firstSeen: '2026-08-10 08:30',
    lastSeen: 'Now',
    latencyMs: 2,
    os: 'Synology DSM 7.2.2',
    services: [
      { port: 22, protocol: 'TCP', service: 'SSH', status: 'Open' },
      { port: 445, protocol: 'TCP', service: 'SMB / CIFS', status: 'Open' },
      { port: 5000, protocol: 'TCP', service: 'HTTP (DSM Admin)', status: 'Open' },
      { port: 5001, protocol: 'TCP', service: 'HTTPS (DSM Secure)', status: 'Open' }
    ],
    history: [
      { id: 'h29', timestamp: '2026-10-03 17:50', type: 'online', description: 'RAID volume healthy' }
    ]
  },
  {
    id: 'dev-24',
    name: 'Generic Android TV Box',
    hostname: 'amlogic-box-tv',
    ip: '192.168.1.115',
    mac: '00:1A:7D:DA:81:45',
    vendor: 'Amlogic, Inc.',
    type: 'Unknown',
    status: 'online',
    firstSeen: '2026-09-10 14:00',
    lastSeen: '14m ago',
    latencyMs: 16,
    services: [
      { port: 5555, protocol: 'TCP', service: 'ADB Daemon (Android Debug)', status: 'Open' }
    ],
    history: [
      { id: 'h30', timestamp: '2026-10-03 17:42', type: 'online', description: 'Port 5555 discovered open' }
    ]
  }
];
