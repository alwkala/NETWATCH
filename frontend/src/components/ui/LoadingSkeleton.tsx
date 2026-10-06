import React from 'react';

export const TableLoadingSkeleton: React.FC = () => {
  return (
    <div className="space-y-3 p-4 animate-pulse">
      {/* Header bar */}
      <div className="h-6 bg-neutral-200 dark:bg-neutral-800 rounded w-1/4 mb-4" />
      {/* Table rows */}
      {[...Array(6)].map((_, i) => (
        <div
          key={i}
          className="h-11 bg-neutral-100 dark:bg-neutral-800/60 rounded flex items-center justify-between px-4 gap-4"
        >
          <div className="w-4 h-4 bg-neutral-200 dark:bg-neutral-700 rounded-full" />
          <div className="w-32 h-3.5 bg-neutral-200 dark:bg-neutral-700 rounded" />
          <div className="w-24 h-3 bg-neutral-200 dark:bg-neutral-700 rounded" />
          <div className="w-28 h-3 bg-neutral-200 dark:bg-neutral-700 rounded" />
          <div className="w-16 h-3 bg-neutral-200 dark:bg-neutral-700 rounded" />
          <div className="w-12 h-3 bg-neutral-200 dark:bg-neutral-700 rounded" />
        </div>
      ))}
    </div>
  );
};

export const DashboardLoadingSkeleton: React.FC = () => {
  return (
    <div className="space-y-6 animate-pulse p-6">
      {/* Header skeleton */}
      <div className="flex justify-between items-center">
        <div className="space-y-2">
          <div className="h-6 w-48 bg-neutral-200 dark:bg-neutral-800 rounded" />
          <div className="h-4 w-32 bg-neutral-100 dark:bg-neutral-800 rounded" />
        </div>
        <div className="h-8 w-28 bg-neutral-200 dark:bg-neutral-800 rounded" />
      </div>

      {/* KPI Cards skeleton */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-20 bg-neutral-100 dark:bg-neutral-800 rounded p-3" />
        ))}
      </div>

      {/* Content skeleton */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 h-72 bg-neutral-100 dark:bg-neutral-800 rounded" />
        <div className="h-72 bg-neutral-100 dark:bg-neutral-800 rounded" />
      </div>
    </div>
  );
};
