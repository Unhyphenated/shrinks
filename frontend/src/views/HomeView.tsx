import { useState, useEffect } from 'react';
import { 
  Copy, 
  Activity, 
  Zap, 
  Globe, 
  Github, 
  Check,
  Command,
  Terminal,
  Cpu,
  Layers,
  Hash
} from 'lucide-react';
import { BentoItem } from '../components/BentoItem';
import { apiClient } from '../api/client';
import { useAuth } from '../hooks/useAuth';
import type { Link, CreateLinkResponse, ViewState } from '../types';

// Placeholder domain - replace when you secure a domain
const SHORT_DOMAIN = 'shrinks.io';

interface HomeViewProps {
  setView: (v: ViewState) => void;
}

export function HomeView({ setView }: HomeViewProps) {
  const [url, setUrl] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [recentLinks, setRecentLinks] = useState<(Link & { short_url: string })[]>([]);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [justCreated, setJustCreated] = useState<CreateLinkResponse | null>(null);
  
  const { isAuthenticated } = useAuth();

  // Fetch user's recent links if authenticated
  useEffect(() => {
    if (isAuthenticated) {
      fetchRecentLinks();
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

  const displayLinks = recentLinks.length > 0 ? recentLinks : [];

  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="flex flex-col items-center text-center mb-20">
        <div className="inline-flex items-center gap-2 px-3 py-1 bg-zinc-50 border border-zinc-200 text-[#E11D48] text-xs font-mono mb-8 font-bold uppercase tracking-wider">
          <Terminal className="w-3 h-3" />
          <span>v2.0.0 Stable</span>
        </div>
        
        <h1 className="text-6xl md:text-8xl font-bold text-zinc-900 tracking-tighter mb-8 max-w-5xl leading-[0.9]">
          INFRASTRUCTURE <br className="hidden md:block" />
          FOR <span className="text-[#E11D48]">MODERN LINKS.</span>
        </h1>
        
        <p className="text-xl text-zinc-500 max-w-2xl mb-12 leading-relaxed font-light">
          The open-source link management platform designed for developers. 
          High-performance redirects, real-time analytics, and API-first design.
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
            <div className="hidden md:flex items-center px-4 bg-white border-l-2 border-zinc-100">
              <div className="px-2 py-1 bg-zinc-100 text-[10px] font-mono text-zinc-500 font-bold border border-zinc-200">CMD + K</div>
            </div>
            <button 
              type="submit"
              disabled={isLoading}
              className="bg-[#E11D48] text-white px-10 py-4 font-bold text-sm uppercase tracking-wider hover:bg-black transition-colors disabled:opacity-50 border-l-2 border-black"
            >
              {isLoading ? 'Processing...' : 'Shorten'}
            </button>
          </form>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-[-1px] mb-24 bg-zinc-200 border border-zinc-200">
        <BentoItem title="Total Requests" value="2.4B" sub="+12% this month" icon={Activity} />
        <BentoItem title="Avg. Latency" value="14ms" sub="Global Edge Network" icon={Zap} />
        <BentoItem title="Active Links" value="850k" sub="Across 12 regions" icon={Globe} />
        <BentoItem title="Uptime" value="99.99%" sub="SLA Guaranteed" icon={Cpu} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-12">
        <div className="lg:col-span-2 space-y-6">
          <div className="flex items-center justify-between border-b-2 border-black pb-4">
            <h3 className="text-zinc-900 font-bold flex items-center gap-3 text-xl">
              <Hash className="w-5 h-5 text-[#E11D48]" />
              {isAuthenticated ? 'Your Recent Links' : 'Latest Deployments'}
            </h3>
            <div className="flex gap-2 text-xs font-mono uppercase">
              <button className="text-white bg-black px-3 py-1">All</button>
              <button className="text-zinc-500 bg-zinc-100 px-3 py-1 hover:bg-zinc-200 border border-zinc-200">Active</button>
            </div>
          </div>
          
          {displayLinks.length === 0 ? (
            <div className="border border-zinc-200 bg-zinc-50 p-8 text-center">
              <p className="text-zinc-500 font-mono text-sm">
                {isAuthenticated 
                  ? 'No links yet. Create your first short link above!'
                  : 'Log in to see your links here.'}
              </p>
            </div>
          ) : (
            <div className="space-y-0 border border-zinc-200 bg-zinc-100 p-[1px] gap-[1px] grid">
              {displayLinks.map((link) => (
                <div key={link.id} className="group flex flex-col md:flex-row items-start md:items-center justify-between p-4 bg-white hover:border-l-[6px] border-l-[6px] border-l-transparent hover:border-l-[#E11D48] transition-all duration-100">
                  <div className="flex items-center gap-4 w-full md:w-auto overflow-hidden">
                    <div className="w-10 h-10 bg-zinc-50 border border-zinc-200 flex items-center justify-center shrink-0 font-mono text-xs text-[#E11D48] font-bold">GET</div>
                    <div className="flex flex-col min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-zinc-900 font-mono font-bold">{link.short_url}</span>
                        <span className="inline-flex items-center px-1.5 py-0.5 text-[10px] font-bold uppercase bg-green-50 text-green-700 border border-green-200">Active</span>
                      </div>
                      <span className="text-zinc-400 text-xs truncate max-w-[200px] md:max-w-md">{link.long_url}</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-6 mt-4 md:mt-0 w-full md:w-auto justify-between md:justify-end">
                    <div className="flex flex-col items-end">
                      <span className="text-zinc-900 font-bold text-sm">—</span>
                      <span className="text-zinc-400 text-[10px] uppercase tracking-wider">Clicks</span>
                    </div>
                    <button 
                      onClick={() => handleCopy(link.short_url, String(link.id))}
                      className="flex items-center gap-2 px-4 py-2 bg-zinc-50 hover:bg-black text-zinc-600 hover:text-white border border-zinc-200 hover:border-black text-xs font-bold uppercase transition-all"
                    >
                      {copiedId === String(link.id) ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                      {copiedId === String(link.id) ? 'Copied' : 'Copy'}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {isAuthenticated && displayLinks.length > 0 && (
            <div className="text-center">
              <button 
                onClick={() => setView('links')}
                className="text-sm font-mono text-[#E11D48] hover:text-black underline underline-offset-4"
              >
                View all links →
              </button>
            </div>
          )}
        </div>

        <div className="space-y-8">
          <div className="bg-zinc-50 border-2 border-zinc-200 p-8">
            <h4 className="text-zinc-900 font-bold mb-6 flex items-center gap-2 uppercase tracking-wide text-sm">
              <Layers className="w-4 h-4 text-[#E11D48]" />
              Integration
            </h4>
            <p className="text-sm text-zinc-500 mb-8 leading-relaxed">
              Connect Shrinks to your existing workflow. We support Vercel, Netlify, and GitHub Actions out of the box.
            </p>
            <div className="space-y-3">
              <div className="flex items-center gap-3 p-4 bg-white border border-zinc-200 hover:border-black cursor-pointer transition-colors group">
                <Github className="w-5 h-5 text-zinc-400 group-hover:text-black" />
                <div className="text-sm text-zinc-900 font-bold">GitHub Action</div>
              </div>
            </div>
          </div>
          <div className="p-8 bg-[#E11D48] text-white relative overflow-hidden group">
            <div className="relative z-10">
              <h4 className="font-bold text-2xl mb-2 font-mono">PRO_PLAN</h4>
              <p className="text-red-100 text-sm mb-6 max-w-[80%]">Get custom domains, extended analytics, and team seats.</p>
              <button className="bg-white text-black px-6 py-3 text-sm font-bold uppercase tracking-wider hover:bg-black hover:text-white transition-colors">Upgrade Now</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
