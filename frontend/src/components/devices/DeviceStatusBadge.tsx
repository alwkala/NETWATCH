import React from 'react';
import { DeviceStatus } from '../../types/device';

interface DeviceStatusBadgeProps {
  status: DeviceStatus;
  isNew?: boolean;
  showText?: boolean;
  className?: string;
}

export const DeviceStatusBadge: React.FC<DeviceStatusBadgeProps> = ({
  status,
  isNew,
  showText = true,
  className = ''
}) => {
  const isOnline = status === 'online';

  return (
    <div className={`inline-flex items-center gap-1.5 text-xs font-medium ${className}`}>
      <span
        className={`w-2 h-2 rounded-full shrink-0 ${
          isOnline
            ? 'bg-emerald-500 shadow-xs ring-2 ring-emerald-500/20'
            : 'bg-neutral-400 dark:bg-neutral-500'
        }`}
        aria-hidden="true"
      />
      {showText && (
        <span
          className={
            isOnline
              ? 'text-emerald-700 dark:text-emerald-400'
              : 'text-neutral-500 dark:text-neutral-400'
          }
        >
          {isOnline ? 'Online' : 'Offline'}
        </span>
      )}
      {isNew && (
        <span className="text-[11px] font-mono tracking-tight text-amber-600 dark:text-amber-400 ml-1">
          · NEW
        </span>
      )}
    </div>
  );
};
