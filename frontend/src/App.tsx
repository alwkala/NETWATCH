import React, { useState } from 'react';
import { ThemeProvider } from './context/ThemeContext';
import { NetworkProvider, useNetwork } from './context/NetworkContext';
import { TitleBar } from './components/layout/TitleBar';
import { Sidebar } from './components/layout/Sidebar';
import { TopBar } from './components/layout/TopBar';
import { Dashboard } from './pages/Dashboard';
import { Devices } from './pages/Devices';
import { DeviceDetails } from './pages/DeviceDetails';
import { Network } from './pages/Network';
import { Scanner } from './pages/Scanner';
import { Events } from './pages/Events';
import { Settings } from './pages/Settings';
import { NetworkService } from './services/NetworkService';

const MainContent: React.FC = () => {
  const { activePage } = useNetwork();

  return (
    <main className="flex-1 overflow-y-auto bg-neutral-100/60 dark:bg-neutral-900/60">
      {(() => {
        switch (activePage) {
          case 'dashboard':
            return <Dashboard />;
          case 'devices':
            return <Devices />;
          case 'device-details':
            return <DeviceDetails />;
          case 'network':
            return <Network />;
          case 'scanner':
            return <Scanner />;
          case 'events':
            return <Events />;
          case 'settings':
            return <Settings />;
          default:
            return <Dashboard />;
        }
      })()}
    </main>
  );
};

const AppShell: React.FC = () => {
  const { service } = useNetwork();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <div className="flex flex-col h-screen w-screen overflow-hidden bg-neutral-100 dark:bg-neutral-950 font-sans text-neutral-900 dark:text-neutral-100">
      {/* Windows 10/11 Title Bar with Controls & Prototype State Switcher (Browser simulation only) */}
      {service.isSimulated && <TitleBar />}

      {/* Application Main Layout */}
      <div className="flex flex-1 overflow-hidden relative">
        {/* Navigation Sidebar (Desktop + Mobile Slide-over Drawer) */}
        <Sidebar
          mobileOpen={mobileMenuOpen}
          onMobileClose={() => setMobileMenuOpen(false)}
        />

        {/* Content Column: TopBar + Page Body */}
        <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
          <TopBar onToggleMobileMenu={() => setMobileMenuOpen(v => !v)} />
          <MainContent />
        </div>
      </div>
    </div>
  );
};

export default function App({ service, startupError }: { service?: NetworkService; startupError?: string }) {
  return (
    <ThemeProvider>
      <NetworkProvider service={service} startupError={startupError}>
        <AppShell />
      </NetworkProvider>
    </ThemeProvider>
  );
}
