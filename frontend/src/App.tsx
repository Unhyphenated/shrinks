import { useState } from 'react';
import { Github, Twitter, ArrowRight } from 'lucide-react';
import { AuthProvider } from './hooks/useAuth';
import { Navbar } from './components/Navbar';
import { HomeView } from './views/HomeView';
import { AnalyticsView } from './views/AnalyticsView';
import { LinksView } from './views/LinksView';
import { LoginView } from './views/LoginView';
import { ForgotPasswordView } from './views/ForgotPasswordView';
import { EmailTemplateView } from './views/EmailTemplateView';
import type { ViewState } from './types';

function AppContent() {
  const [view, setView] = useState<ViewState>('home');
  const [selectedLinkCode, setSelectedLinkCode] = useState<string | null>(null);

  // Reset selected link when navigating away from analytics
  const handleSetView = (newView: ViewState) => {
    if (newView !== 'analytics') {
      setSelectedLinkCode(null);
    }
    setView(newView);
  };

  return (
    <div className="min-h-screen bg-white text-zinc-900 font-sans selection:bg-[#E11D48] selection:text-white">
      {/* Grid Background */}
      <div 
        className="fixed inset-0 pointer-events-none" 
        style={{ 
          backgroundImage: 'radial-gradient(circle at 1px 1px, #e4e4e7 1px, transparent 0)', 
          backgroundSize: '32px 32px' 
        }}
      />

      <Navbar currentView={view} setView={handleSetView} />

      <main className="relative z-10 pt-32 pb-20 max-w-7xl mx-auto px-6">
        {view === 'home' && <HomeView setView={handleSetView} />}
        {view === 'analytics' && <AnalyticsView selectedLinkCode={selectedLinkCode} />}
        {view === 'login' && <LoginView setView={handleSetView} />}
        {view === 'forgot-password' && <ForgotPasswordView setView={handleSetView} />}
        {view === 'email-preview' && <EmailTemplateView />}
        {view === 'links' && (
          <LinksView 
            setView={handleSetView} 
            setSelectedLinkCode={setSelectedLinkCode} 
          />
        )}
      </main>

      <footer className="border-t-2 border-zinc-200 bg-zinc-50 pt-20 pb-10 mt-auto">
        <div className="max-w-7xl mx-auto px-6 grid grid-cols-2 md:grid-cols-4 gap-12 mb-16">
          <div>
            <h5 className="text-black font-bold mb-6 uppercase text-xs tracking-wider">Platform</h5>
            <ul className="space-y-3 text-sm text-zinc-500 font-medium">
              <li><button onClick={() => handleSetView('home')} className="hover:text-[#E11D48]">Shortener</button></li>
              <li><button onClick={() => handleSetView('analytics')} className="hover:text-[#E11D48]">Analytics</button></li>
            </ul>
          </div>
          <div>
            <h5 className="text-black font-bold mb-6 uppercase text-xs tracking-wider">Company</h5>
            <ul className="space-y-3 text-sm text-zinc-500 font-medium">
              <li><a href="#" className="hover:text-[#E11D48]">About</a></li>
              <li><a href="#" className="hover:text-[#E11D48]">Careers</a></li>
            </ul>
          </div>
          <div>
            <h5 className="text-black font-bold mb-6 uppercase text-xs tracking-wider">Legal</h5>
            <ul className="space-y-3 text-sm text-zinc-500 font-medium">
              <li><a href="#" className="hover:text-[#E11D48]">Privacy</a></li>
              <li><a href="#" className="hover:text-[#E11D48]">Terms</a></li>
            </ul>
          </div>
          <div className="col-span-2 md:col-span-1">
            <h5 className="text-black font-bold mb-6 uppercase text-xs tracking-wider">Subscribe</h5>
            <p className="text-xs text-zinc-500 mb-4">Latest updates on our API and features.</p>
            <div className="flex border-2 border-zinc-200 bg-white p-1">
              <input 
                type="email" 
                placeholder="email@company.com" 
                className="flex-1 bg-transparent border-none px-3 py-2 text-sm text-black w-full focus:ring-0 placeholder-zinc-400 outline-none" 
              />
              <button className="bg-black text-white px-4 py-2 text-sm font-bold hover:bg-[#E11D48] transition-colors">
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
        <div className="max-w-7xl mx-auto px-6 pt-8 border-t-2 border-zinc-200 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <div className="w-5 h-5 bg-[#E11D48] flex items-center justify-center">
              <span className="font-mono text-white font-bold text-[10px]">/</span>
            </div>
            <span className="text-sm text-zinc-500 font-medium">© 2024 Shrinks Inc.</span>
          </div>
          <div className="flex gap-6 text-zinc-400">
            <Twitter className="w-5 h-5 hover:text-black cursor-pointer transition-colors" />
            <Github className="w-5 h-5 hover:text-black cursor-pointer transition-colors" />
          </div>
        </div>
      </footer>
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}

export default App;
