import { useState, useEffect } from 'react';
import { 
  Copy, 
  Activity, 
  Globe, 
  Check,
  Command,
  Terminal,
} from 'lucide-react';
import { BentoItem } from '../components/BentoItem';
import { apiClient } from '../api/client';
import { useAuth } from '../hooks/useAuth';
import type { Link, CreateLinkResponse } from '../types';

// Placeholder domain - replace when you secure a domain
const SHORT_DOMAIN = 'shrinks.io';

export function HomeView() {
  const [url, setUrl] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [, setRecentLinks] = useState<(Link & { short_url: string })[]>([]);
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
        <BentoItem title="Total Requests" value="2.4B" icon={Activity} />
        <BentoItem title="Active Links" value="850k" icon={Globe} />
      </div>

    </div>
  );
}
