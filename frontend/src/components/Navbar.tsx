import { useAuth } from '../hooks/useAuth';
import type { ViewState } from '../types';

interface NavbarProps {
  currentView: ViewState;
  setView: (v: ViewState) => void;
}

export function Navbar({ currentView, setView }: NavbarProps) {
  const { isAuthenticated, user, logout } = useAuth();

  const handleLogout = async () => {
    await logout();
    setView('home');
  };

  return (
    <nav className="fixed top-0 w-full z-50 border-b-2 border-zinc-200 bg-white/95">
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        <div 
          className="flex items-center gap-2 cursor-pointer group"
          onClick={() => setView('home')}
        >
          <div className="w-6 h-6 bg-[#E11D48] flex items-center justify-center transition-transform group-hover:rotate-180">
            <span className="font-mono text-white font-bold text-xs">/</span>
          </div>
          <span className="text-lg font-bold tracking-tight text-zinc-900 group-hover:text-[#E11D48] transition-colors">
            shrinks<span className="text-[#E11D48]">.</span>
          </span>
        </div>

        <div className="hidden md:flex items-center gap-8 text-sm font-medium text-zinc-500 font-mono uppercase tracking-wider">
          <button 
            onClick={() => setView('home')}
            className={`hover:text-black transition-colors ${currentView === 'home' ? 'text-black font-bold underline decoration-2 underline-offset-4 decoration-[#E11D48]' : ''}`}
          >
            Shortener
          </button>
          {isAuthenticated && (
            <>
              <button 
                onClick={() => setView('analytics')}
                className={`hover:text-black transition-colors ${currentView === 'analytics' ? 'text-black font-bold underline decoration-2 underline-offset-4 decoration-[#E11D48]' : ''}`}
              >
                Analytics
              </button>
              <button 
                onClick={() => setView('links')}
                className={`hover:text-black transition-colors ${currentView === 'links' ? 'text-black font-bold underline decoration-2 underline-offset-4 decoration-[#E11D48]' : ''}`}
              >
                My Links
              </button>
            </>
          )}
        </div>

        <div className="flex items-center gap-4">
          {isAuthenticated ? (
            <>
              <span className="hidden md:block text-sm text-zinc-500 font-mono">
                {user?.email}
              </span>
              <button 
                onClick={handleLogout}
                className="hidden md:block text-zinc-500 hover:text-black text-sm font-medium font-mono uppercase transition-colors"
              >
                Log out
              </button>
            </>
          ) : (
            currentView !== 'login' && (
              <button 
                onClick={() => setView('login')}
                className="hidden md:block text-zinc-500 hover:text-black text-sm font-medium font-mono uppercase transition-colors"
              >
                Log in
              </button>
            )
          )}
          <button className="bg-black text-white px-5 py-2 text-sm font-bold hover:bg-[#E11D48] transition-colors border-l-2 border-white hover:border-black uppercase tracking-wider">
            Console
          </button>
        </div>
      </div>
    </nav>
  );
}
