import { useState } from "react";
import { Github, Linkedin } from "lucide-react";
import { AuthProvider } from "./hooks/useAuth";
import { Navbar } from "./components/Navbar";
import { HomeView } from "./views/HomeView";
import { AnalyticsView } from "./views/AnalyticsView";
import { LinksView } from "./views/LinksView";
import { LoginView } from "./views/LoginView";
import { ForgotPasswordView } from "./views/ForgotPasswordView";
import { EmailTemplateView } from "./views/EmailTemplateView";
import type { ViewState } from "./types";

function AppContent() {
  const [view, setView] = useState<ViewState>("home");
  const [selectedLinkCode, setSelectedLinkCode] = useState<string | null>(null);

  // Reset selected link when navigating away from analytics
  const handleSetView = (newView: ViewState) => {
    if (newView !== "analytics") {
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
          backgroundImage:
            "radial-gradient(circle at 1px 1px, #e4e4e7 1px, transparent 0)",
          backgroundSize: "32px 32px",
        }}
      />

      <Navbar currentView={view} setView={handleSetView} />

      <main className="relative z-10 pt-32 pb-20 max-w-7xl mx-auto px-6">
        {view === "home" && (
          <HomeView
            setView={handleSetView}
            setSelectedLinkCode={setSelectedLinkCode}
          />
        )}
        {view === "analytics" && (
          <AnalyticsView
            selectedLinkCode={selectedLinkCode}
            setView={handleSetView}
          />
        )}
        {view === "login" && <LoginView setView={handleSetView} />}
        {view === "forgot-password" && (
          <ForgotPasswordView setView={handleSetView} />
        )}
        {view === "email-preview" && <EmailTemplateView />}
        {view === "links" && (
          <LinksView
            setView={handleSetView}
            setSelectedLinkCode={setSelectedLinkCode}
          />
        )}
      </main>

      <footer className="relative border-t-2 border-zinc-200 bg-zinc-50 py-6 mt-auto overflow-hidden">
        <div className="max-w-7xl mx-auto px-6 relative z-10">
          <div className="flex flex-col md:flex-row items-center justify-between gap-8">
            {/* Left: Shrinks logo + copyright */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 bg-[#E11D48] flex items-center justify-center">
                <span className="font-mono text-white font-bold text-sm">
                  /
                </span>
              </div>
              <span className="text-sm text-zinc-500 font-medium">
                © 2026 Shrinks
              </span>
            </div>

            {/* Right: Social links + Made by */}
            <div className="flex flex-col items-center md:items-end gap-3">
              <div className="flex items-center gap-4">
                <a
                  href="https://github.com/Unhyphenated"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-zinc-400 hover:text-black transition-colors cursor-pointer"
                >
                  <Github className="w-5 h-5" />
                </a>
                <a
                  href="https://linkedin.com/in/julian-jong"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-zinc-400 hover:text-black transition-colors cursor-pointer"
                >
                  <Linkedin className="w-5 h-5" />
                </a>
                <a
                  href="https://unhyphenated.com"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-zinc-400 hover:text-black transition-all cursor-pointer font-bold text-lg tracking-tight"
                  style={{ fontFamily: "system-ui, -apple-system, sans-serif" }}
                >
                  un-
                </a>
              </div>
              <p className="text-xs text-zinc-400 font-mono">
                Made by Julian Jong{" "}
                <span className="text-zinc-300">(Unhyphenated)</span>
              </p>
            </div>
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
