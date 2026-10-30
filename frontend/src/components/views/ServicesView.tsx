import React, { useState } from 'react';
import { Cpu, RefreshCw } from 'lucide-react';
import { ServiceItem } from '../../types';

interface ServicesViewProps {
  services: ServiceItem[];
  onRestartService: (name: string) => Promise<void>;
}

export const ServicesView: React.FC<ServicesViewProps> = ({ services, onRestartService }) => {
  const [loadingName, setLoadingName] = useState<string | null>(null);

  const handleRestart = async (name: string) => {
    setLoadingName(name);
    try {
      await onRestartService(name);
    } finally {
      setLoadingName(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">System Services & Daemons</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Monitor and control systemd service units powering web serving, runtimes, and databases.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {services.map((service) => {
          const isLoading = loadingName === service.name;
          return (
            <div
              key={service.name}
              className="bg-white border border-slate-200 rounded-lg p-5 flex flex-col justify-between"
            >
              <div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Cpu className="w-4 h-4 text-blue-600" />
                    <h3 className="font-bold text-slate-900 text-sm">{service.display_name}</h3>
                  </div>
                  <span className="font-mono text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded text-[10px] border border-emerald-200 uppercase font-semibold">
                    {service.status}
                  </span>
                </div>
                <div className="text-slate-500 font-mono text-xs mt-2">
                  unit: {service.name}.service
                </div>
              </div>

              <div className="mt-5 pt-3 border-t border-slate-100 flex items-center justify-end gap-2">
                <button
                  onClick={() => handleRestart(service.name)}
                  disabled={isLoading}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded border border-slate-200 hover:bg-slate-50 text-slate-700 transition-colors disabled:opacity-50"
                >
                  <RefreshCw className={`w-3.5 h-3.5 text-slate-500 ${isLoading ? 'animate-spin' : ''}`} />
                  <span>Restart</span>
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
