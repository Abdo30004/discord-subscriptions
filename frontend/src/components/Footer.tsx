import React from 'react';
import { Server, Activity, ShieldCheck, Cpu } from 'lucide-react';

export default function Footer() {
  const services = [
    { name: 'auth-svc', port: 8080, status: 'online' },
    { name: 'catalog-svc', port: 8081, status: 'online' },
    { name: 'billing-svc', port: 8082, status: 'online' },
    { name: 'deploy-svc', port: 8083, status: 'online' },
    { name: 'monitor-svc', port: 8084, status: 'online' },
  ];

  return (
    <footer className="w-full border-t border-card-border bg-[#07090D] py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Service Health Fleet Banner */}
        <div className="p-4 rounded-xl bg-card border border-card-border">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <Server className="w-4 h-4 text-blurple" />
              <span className="text-xs font-mono font-medium text-slate-300">
                Microservices Mesh Status:
              </span>
            </div>
            <div className="flex flex-wrap items-center gap-3">
              {services.map((svc) => (
                <div
                  key={svc.name}
                  className="flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-slate-900/80 border border-slate-800 text-[11px] font-mono text-slate-300"
                >
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                  <span className="text-white font-semibold">{svc.name}</span>
                  <span className="text-slate-500">:{svc.port}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Footer Details */}
        <div className="flex flex-col sm:flex-row items-center justify-between text-xs text-slate-500 gap-4">
          <p>© 2026 Discord Subscriptions Cloud Platform. All rights reserved.</p>
          <div className="flex items-center gap-6">
            <span className="flex items-center gap-1 text-slate-400">
              <ShieldCheck className="w-3.5 h-3.5 text-blurple" /> HashiCorp Vault Encrypted
            </span>
            <span className="flex items-center gap-1 text-slate-400">
              <Cpu className="w-3.5 h-3.5 text-emerald-400" /> Kubernetes Orchestrated
            </span>
          </div>
        </div>
      </div>
    </footer>
  );
}
