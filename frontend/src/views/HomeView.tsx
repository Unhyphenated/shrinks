import { useState, useEffect } from 'react';
import { 
  Copy, 
  Activity, 
  Globe, 
  Check,
  Command,
  Terminal,
  BarChart2,
} from 'lucide-react';
import { BentoItem } from '../components/BentoItem';
import { apiClient } from '../api/client';
import { useAuth } from '../hooks/useAuth';
import type { Link, CreateLinkResponse, GlobalStats, ViewState } from '../types';

// Placeholder domain - replace when you secure a domain
const SHORT_DOMAIN = 'shrinks.io';

interface HomeViewProps {
  setView: (v: ViewState) => void;
  setSelectedLinkCode: (code: string) => void;
}

export function HomeView({ setView, setSelectedLinkCode }: HomeViewProps) {
  const [url, setUrl] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [recentLinks, setRecentLinks] = useState<(Link & { short_url: string })[]>([]);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [justCreated, setJustCreated] = useState<CreateLinkResponse | null>(null);
  const [globalStats, setGlobalStats] = useState<GlobalStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);
  
  const { isAuthenticated } = useAuth();

  // Fetch global stats on mount
  useEffect(() => {
    fetchGlobalStats();
  }, []);

  const fetchGlobalStats = async () => {
    setStatsLoading(true);
    try {
      const stats = await apiClient.getGlobalStats();
      setGlobalStats(stats);
    } catch (err) {
      console.error('Failed to fetch global stats:', err);
      // Use fallback data if fetch fails
      setGlobalStats({ total_links: 0, total_requests: 0 });
    } finally {
      setStatsLoading(false);
    }
  };

  // Fetch user's recent links if authenticated
  useEffect(() => {
    if (isAuthenticated) {
      fetchRecentLinks();
    } else {
      setRecentLinks([]);
    }
  }, [isAuthenticated]);

  const fetchRecentLinks = async () => {
    try {
      const response = await apiClient.getLinks(5, 0);
      const linksWithShortUrl = response.links.map(link => ({
        ...link,
        short_url: `${SHORT_DOMAIN}/${link.short_code}`
      }));
      setRecentLinks(linksWithShortUrl);
    } catch (err) {
      console.error('Failed to fetch links:', err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url) return;
    
    setIsLoading(true);
    setError(null);
    setJustCreated(null);
    
    try {
      const response = await apiClient.shortenUrl(url);
      setJustCreated(response);
      setUrl('');
      
      // If authenticated, refresh the links list
      if (isAuthenticated) {
        await fetchRecentLinks();
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to shorten URL';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleViewAnalytics = (shortCode: string) => {
    setSelectedLinkCode(shortCode);
    setView('analytics');
  };

  const formatNumber = (num: number): string => {
    if (num >= 1_000_000_000) {
      return (num / 1_000_000_000).toFixed(1) + 'B';
    }
    if (num >= 1_000_000) {
      return (num / 1_000_000).toFixed(1) + 'M';
    }
    if (num >= 1_000) {
      return (num / 1_000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="flex flex-col items-center text-center mb-20">
        <div className="inline-flex items-center gap-2 px-3 py-1 bg-zinc-50 border border-zinc-200 text-[#E11D48] text-xs font-mono mb-8 font-bold uppercase tracking-wider">
          <Terminal className="w-3 h-3" />
          <span>v1.0.0</span>
        </div>
        
        <h1 className="text-6xl md:text-8xl font-bold text-zinc-900 tracking-tighter mb-8 max-w-5xl leading-[0.9]">
          SHORTEN LINKS, <br className="hidden md:block" />
          <span className="text-[#E11D48]">NOT YOUR PATIENCE.</span>
        </h1>
        
        <p className="text-xl text-zinc-500 max-w-2xl mb-12 leading-relaxed font-light">
          A minimal, high-performance link shortener. Shorten URLs, track clicks, and manage your links — all in one place.
        </p>

        <div className="w-full max-w-3xl">
          {error && (
            <div className="mb-4 p-3 bg-red-50 border-2 border-red-200 text-red-700 text-sm font-mono text-left">
              {error}
            </div>
          )}
          
          {justCreated && (
            <div className="mb-4 p-4 bg-green-50 border-2 border-green-200 text-left">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-xs font-mono text-green-600 uppercase font-bold mb-1">Link Created!</p>
                  <p className="font-mono font-bold text-green-900">{SHORT_DOMAIN}/{justCreated.short_code}</p>
                </div>
                <button
                  onClick={() => handleCopy(`${SHORT_DOMAIN}/${justCreated.short_code}`, 'new')}
                  className="flex items-center gap-2 px-4 py-2 bg-green-600 hover:bg-green-700 text-white text-xs font-bold uppercase transition-colors"
                >
                  {copiedId === 'new' ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                  {copiedId === 'new' ? 'Copied' : 'Copy'}
                </button>
              </div>
            </div>
          )}

          <form onSubmit={handleSubmit} className="relative flex items-stretch shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] border-2 border-black bg-white transition-transform active:translate-x-1 active:translate-y-1 active:shadow-none">
            <div className="pl-6 pr-4 flex items-center justify-center bg-zinc-50 border-r-2 border-black">
              <Command className="w-5 h-5 text-zinc-400" />
            </div>
            <input 
              type="url" 
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="Paste your URL here..." 
              className="flex-1 bg-white border-none text-zinc-900 placeholder-zinc-400 focus:ring-0 text-lg font-mono py-6 px-6 outline-none"
              disabled={isLoading}
            />
            <button 
              type="submit"
              disabled={isLoading}
              className="bg-[#E11D48] text-white px-10 py-4 font-bold text-sm uppercase tracking-wider hover:bg-black transition-colors disabled:opacity-50 border-l-2 border-black cursor-pointer"
            >
              {isLoading ? 'Processing...' : 'Shorten'}
            </button>
          </form>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-[-1px] mb-24 bg-zinc-200 border border-zinc-200">
        <BentoItem 
          title="Total Requests" 
          value={statsLoading ? '...' : formatNumber(globalStats?.total_requests || 0)} 
          icon={Activity} 
        />
        <BentoItem 
          title="Active Links" 
          value={statsLoading ? '...' : formatNumber(globalStats?.total_links || 0)} 
          icon={Globe} 
        />
      </div>

      {/* Recent Links Section - Only show if authenticated and has links */}
      {isAuthenticated && recentLinks.length > 0 && (
        <div className="mb-24">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-2xl font-bold text-zinc-900 uppercase tracking-tighter">
              Recent <span className="text-[#E11D48]">Links</span>
            </h3>
            <button
              onClick={() => setView('links')}
              className="text-sm font-mono text-zinc-500 hover:text-black transition-colors uppercase"
            >
              View All →
            </button>
          </div>
          
          <div className="grid grid-cols-1 gap-3">
            {recentLinks.map((link) => (
              <div
                key={link.id}
                className="bg-white border-2 border-zinc-200 p-4 hover:border-black transition-colors group"
              >
                <div className="flex items-center justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="font-bold font-mono text-sm text-zinc-900 mb-1">
                      {SHORT_DOMAIN}/{link.short_code}
                    </div>
                    <div className="text-xs text-zinc-400 truncate group-hover:text-[#E11D48] transition-colors">
                      {link.long_url}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleViewAnalytics(link.short_code)}
                      className="flex items-center gap-2 px-3 py-2 bg-zinc-100 hover:bg-black hover:text-white border-2 border-zinc-200 hover:border-black text-xs font-bold uppercase transition-colors"
                      title="View Analytics"
                    >
                      <BarChart2 className="w-3 h-3" />
                      <span className="hidden sm:inline">Analytics</span>
                    </button>
                    <button
                      onClick={() => handleCopy(link.short_url, link.short_code)}
                      className="p-2 border-2 border-zinc-200 bg-white hover:bg-zinc-100 transition-colors"
                      title="Copy Link"
                    >
                      {copiedId === link.short_code ? (
                        <Check className="w-4 h-4 text-green-600" />
                      ) : (
                        <Copy className="w-4 h-4 text-zinc-400" />
                      )}
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

    </div>
  );
}
