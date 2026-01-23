import { useState, useEffect } from 'react';
import { 
  Activity, 
  Zap, 
  Globe, 
  BarChart2,
  Calendar,
  MousePointer,
  Server,
  Link as LinkIcon,
  Smartphone,
  Monitor,
  Tablet
} from 'lucide-react';
import { BentoItem } from '../components/BentoItem';
import { BarChart } from '../components/BarChart';
import { apiClient } from '../api/client';
import type { AnalyticsSummary } from '../types';

// Placeholder domain
const SHORT_DOMAIN = 'shrinks.io';

interface AnalyticsViewProps {
  selectedLinkCode: string | null;
}

type Period = '24h' | '7d' | '30d';

export function AnalyticsView({ selectedLinkCode }: AnalyticsViewProps) {
  const [analytics, setAnalytics] = useState<AnalyticsSummary | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [period, setPeriod] = useState<Period>('7d');

  const fetchAnalytics = async () => {
    if (!selectedLinkCode) return;
    
    setIsLoading(true);
    setError(null);
    try {
      const data = await apiClient.getAnalytics(selectedLinkCode, period);
      setAnalytics(data);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch analytics';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (selectedLinkCode) {
      fetchAnalytics();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedLinkCode, period]);

  const getDeviceIcon = (device: string) => {
    const deviceLower = device.toLowerCase();
    if (deviceLower.includes('mobile') || deviceLower.includes('phone')) {
      return Smartphone;
    }
    if (deviceLower.includes('tablet')) {
      return Tablet;
    }
    return Monitor;
  };

  // Calculate top regions from analytics (placeholder - backend doesn't provide this yet)
  const topRegions = [
    { country: 'United States', percentage: 45 },
    { country: 'Germany', percentage: 22 },
    { country: 'Japan', percentage: 15 },
  ];

  if (!selectedLinkCode) {
    return (
      <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div className="flex flex-col items-center justify-center py-20">
          <div className="w-16 h-16 bg-zinc-100 border-2 border-zinc-200 flex items-center justify-center mb-6">
            <BarChart2 className="w-8 h-8 text-zinc-400" />
          </div>
          <h2 className="text-2xl font-bold text-zinc-900 mb-2">No Link Selected</h2>
          <p className="text-zinc-500 font-mono text-sm">
            Select a link from "My Links" to view its analytics.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="flex flex-col md:flex-row md:items-end justify-between mb-12 gap-6">
        <div>
          <div className="inline-flex items-center gap-2 px-3 py-1 bg-black text-white text-xs font-mono mb-4 font-bold uppercase tracking-wider">
            <Activity className="w-3 h-3" />
            <span>Link Analytics</span>
          </div>
          <h2 className="text-5xl font-bold text-zinc-900 tracking-tighter leading-[0.9]">
            TRAFFIC <span className="text-[#E11D48]">INSIGHTS</span>
          </h2>
          <p className="text-zinc-500 font-mono mt-2 flex items-center gap-2">
            <LinkIcon className="w-4 h-4" />
            Showing data for: <span className="text-black font-bold">{SHORT_DOMAIN}/{selectedLinkCode}</span>
          </p>
        </div>
        
        <div className="flex items-center gap-4">
          <div className="flex items-center bg-white border-2 border-zinc-200">
            {(['24h', '7d', '30d'] as Period[]).map((p) => (
              <button
                key={p}
                onClick={() => setPeriod(p)}
                className={`px-4 py-2 text-xs font-bold uppercase border-r border-zinc-200 last:border-r-0 transition-colors ${
                  period === p 
                    ? 'bg-black text-white' 
                    : 'hover:bg-zinc-50 text-zinc-500'
                }`}
              >
                {p}
              </button>
            ))}
          </div>
          <button className="p-2 border-2 border-zinc-200 hover:border-black transition-colors">
            <Calendar className="w-5 h-5" />
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border-2 border-red-200 text-red-700 font-mono text-sm">
          {error}
        </div>
      )}

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-[-1px] bg-zinc-200 border border-zinc-200 mb-12">
        <BentoItem 
          title="Total Clicks" 
          value={isLoading ? '...' : (analytics?.total_clicks?.toLocaleString() || '0')} 
          icon={MousePointer} 
        />
        <BentoItem 
          title="Unique Visitors" 
          value={isLoading ? '...' : (analytics?.unique_visitors?.toLocaleString() || '0')} 
          icon={Activity} 
        />
        <BentoItem 
          title="Avg. Per Day" 
          value={isLoading ? '...' : calculateAvgPerDay(analytics, period)} 
          icon={Zap} 
        />
        <BentoItem 
          title="Bot Traffic" 
          value="<1%" 
          icon={Server} 
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-12">
        {/* Clicks Over Time Chart */}
        <div className="lg:col-span-2 bg-white border-2 border-zinc-200 p-8">
          <div className="flex items-center justify-between mb-8">
            <h3 className="font-bold text-lg uppercase tracking-wider flex items-center gap-2">
              <BarChart2 className="w-5 h-5 text-[#E11D48]" />
              Clicks Over Time
            </h3>
            <span className="text-xs font-mono text-zinc-400">UTC Timezone</span>
          </div>
          <BarChart 
            data={analytics?.clicks_by_date} 
            isLoading={isLoading}
          />
        </div>

        {/* Top Regions */}
        <div className="bg-zinc-900 text-white p-8 relative overflow-hidden">
          <div className="relative z-10 h-full flex flex-col">
            <h3 className="font-bold text-lg uppercase tracking-wider mb-8 flex items-center gap-2">
              <Globe className="w-5 h-5 text-[#E11D48]" />
              Top Regions
            </h3>
            <div className="space-y-6 flex-1">
              {topRegions.map((item, i) => (
                <div key={i} className="group">
                  <div className="flex justify-between text-sm font-mono mb-2">
                    <span className="text-zinc-400">{item.country}</span>
                    <span className="font-bold">{item.percentage}%</span>
                  </div>
                  <div className="w-full h-1 bg-zinc-800">
                    <div className="h-full bg-[#E11D48]" style={{ width: `${item.percentage}%` }}></div>
                  </div>
                </div>
              ))}
            </div>
            <p className="text-xs text-zinc-500 mt-4 font-mono">
              * Region data coming soon
            </p>
          </div>
        </div>
      </div>

      {/* Device & Browser Breakdown */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {/* By Device */}
        <div className="bg-white border-2 border-zinc-200 p-8">
          <h3 className="font-bold text-lg uppercase tracking-wider mb-6 flex items-center gap-2">
            <Smartphone className="w-5 h-5 text-[#E11D48]" />
            By Device
          </h3>
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="animate-pulse">
                  <div className="h-4 bg-zinc-200 rounded w-1/2 mb-2"></div>
                  <div className="h-2 bg-zinc-100 rounded w-full"></div>
                </div>
              ))}
            </div>
          ) : analytics?.clicks_by_device && analytics.clicks_by_device.length > 0 ? (
            <div className="space-y-4">
              {analytics.clicks_by_device.map((item, i) => {
                const total = analytics.clicks_by_device.reduce((sum, d) => sum + d.clicks, 0);
                const percentage = total > 0 ? Math.round((item.clicks / total) * 100) : 0;
                const DeviceIcon = getDeviceIcon(item.device);
                return (
                  <div key={i}>
                    <div className="flex items-center justify-between text-sm mb-2">
                      <div className="flex items-center gap-2">
                        <DeviceIcon className="w-4 h-4 text-zinc-400" />
                        <span className="font-mono">{item.device || 'Unknown'}</span>
                      </div>
                      <span className="font-bold">{percentage}%</span>
                    </div>
                    <div className="w-full h-2 bg-zinc-100">
                      <div className="h-full bg-zinc-900" style={{ width: `${percentage}%` }}></div>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <p className="text-zinc-500 font-mono text-sm">No device data available</p>
          )}
        </div>

        {/* By Browser */}
        <div className="bg-white border-2 border-zinc-200 p-8">
          <h3 className="font-bold text-lg uppercase tracking-wider mb-6 flex items-center gap-2">
            <Globe className="w-5 h-5 text-[#E11D48]" />
            By Browser
          </h3>
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="animate-pulse">
                  <div className="h-4 bg-zinc-200 rounded w-1/2 mb-2"></div>
                  <div className="h-2 bg-zinc-100 rounded w-full"></div>
                </div>
              ))}
            </div>
          ) : analytics?.clicks_by_browser && analytics.clicks_by_browser.length > 0 ? (
            <div className="space-y-4">
              {analytics.clicks_by_browser.map((item, i) => {
                const total = analytics.clicks_by_browser.reduce((sum, b) => sum + b.clicks, 0);
                const percentage = total > 0 ? Math.round((item.clicks / total) * 100) : 0;
                return (
                  <div key={i}>
                    <div className="flex items-center justify-between text-sm mb-2">
                      <span className="font-mono">{item.browser || 'Unknown'}</span>
                      <span className="font-bold">{percentage}%</span>
                    </div>
                    <div className="w-full h-2 bg-zinc-100">
                      <div className="h-full bg-[#E11D48]" style={{ width: `${percentage}%` }}></div>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <p className="text-zinc-500 font-mono text-sm">No browser data available</p>
          )}
        </div>
      </div>

      {/* By OS */}
      {analytics?.clicks_by_os && analytics.clicks_by_os.length > 0 && (
        <div className="mt-8 bg-white border-2 border-zinc-200 p-8">
          <h3 className="font-bold text-lg uppercase tracking-wider mb-6 flex items-center gap-2">
            <Monitor className="w-5 h-5 text-[#E11D48]" />
            By Operating System
          </h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {analytics.clicks_by_os.map((item, i) => {
              const total = analytics.clicks_by_os.reduce((sum, o) => sum + o.clicks, 0);
              const percentage = total > 0 ? Math.round((item.clicks / total) * 100) : 0;
              return (
                <div key={i} className="p-4 bg-zinc-50 border border-zinc-200">
                  <div className="text-2xl font-bold font-mono mb-1">{percentage}%</div>
                  <div className="text-xs text-zinc-500 font-mono uppercase">{item.os || 'Unknown'}</div>
                  <div className="text-sm text-zinc-700 mt-1">{item.clicks} clicks</div>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

function calculateAvgPerDay(analytics: AnalyticsSummary | null, period: Period): string {
  if (!analytics || !analytics.total_clicks) return '0';
  
  const days = period === '24h' ? 1 : period === '7d' ? 7 : 30;
  const avg = Math.round(analytics.total_clicks / days);
  return avg.toLocaleString();
}
