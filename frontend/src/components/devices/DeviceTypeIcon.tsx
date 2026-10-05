import React from 'react';
import { DeviceType } from '../../types/device';
import {
  Router as RouterIcon,
  Laptop,
  Smartphone,
  Tablet as TabletIcon,
  Tv,
  Printer,
  Camera,
  Cpu,
  Server,
  Network,
  Gamepad2,
  HelpCircle
} from 'lucide-react';

interface DeviceTypeIconProps {
  type: DeviceType;
  className?: string;
}

export const DeviceTypeIcon: React.FC<DeviceTypeIconProps> = ({ type, className = 'w-4 h-4' }) => {
  switch (type) {
    case 'Router':
      return <RouterIcon className={className} />;
    case 'Computer':
      return <Laptop className={className} />;
    case 'Phone':
      return <Smartphone className={className} />;
    case 'Tablet':
      return <TabletIcon className={className} />;
    case 'TV':
      return <Tv className={className} />;
    case 'Printer':
      return <Printer className={className} />;
    case 'Camera':
      return <Camera className={className} />;
    case 'IoT':
      return <Cpu className={className} />;
    case 'Server':
      return <Server className={className} />;
    case 'Network Device':
      return <Network className={className} />;
    case 'Game Console':
      return <Gamepad2 className={className} />;
    case 'Unknown':
    default:
      return <HelpCircle className={className} />;
  }
};

