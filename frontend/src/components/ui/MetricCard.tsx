import React from 'react';

interface MetricCardProps {
  label: string;
  value: string | number;
  subtext?: string;
  badge?: string;
  highlight?: 'default' | 'online' | 'offline' | 'new' | 'accent';
  onClick?: () => void;
}

export const MetricCard: React.FC<MetricCardProps> = ({
  label,
  value,
  subtext,
  badge,
  highlight = 'default',
  onClick
}) => {
  const getHighlightStyles = () => {
    switch (highlight) {
      case 'online':
        return 'border-l-2 border-l-emerald-500';
      case 'offline':
        return 'border-l-2 border-l-neutral-400 dark:border-l-neutral-600';
      case 'new':
        return 'border-l-2 border-l-amber-500';
      case 'accent':
        return 'border-l-2 border-l-sky-500';
      default:
        return 'border-l-2 border-l-neutral-300 dark:border-l-neutral-700';
    }
  };

  return (
    <div
      onClick={onClick}
      className={`bg-white dark:bg-neutral-800/90 border border-neutral-200 dark:border-neutral-700/70 rounded-md p-3.5 transition-colors ${getHighlightStyles()} ${
        onClick ? 'cursor-pointer hover:bg-neutral-50 dark:hover:bg-neutral-800' : ''
      }`}
    >
      <div className="flex items-center justify-between gap-1 mb-1">
        <span className="text-[11px] font-semibold tracking-wider text-neutral-500 dark:text-neutral-400 uppercase">
          {label}
        </span>
        {badge && (
          <span className="text-[10px] font-mono text-neutral-500 dark:text-neutral-400">
            {badge}
          </span>
        )}
      </div>
      <div className="flex items-baseline gap-2">
        <span className="text-2xl font-bold tracking-tight text-neutral-900 dark:text-neutral-100 font-mono tabular-nums">
          {value}
        </span>
        {subtext && (
          <span className="text-xs text-neutral-500 dark:text-neutral-400">
            {subtext}
          </span>
        )}
      </div>
    </div>
  );
};
