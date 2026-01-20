import { useState } from 'react';
import { Lock, Mail, ShieldCheck, Eye, EyeOff, ArrowRight, Github } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import type { ViewState } from '../types';

const GoogleIcon = () => (
  <svg viewBox="0 0 24 24" className="w-5 h-5" aria-hidden="true">
    <path
      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
      fill="#4285F4"
    />
    <path
      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      fill="#34A853"
    />
    <path
      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.26-.19-.58z"
      fill="#FBBC05"
    />
    <path
      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      fill="#EA4335"
    />
  </svg>
);

interface LoginViewProps {
  setView: (v: ViewState) => void;
}

export function LoginView({ setView }: LoginViewProps) {
  const [showPassword, setShowPassword] = useState(false);
  const [isRegister, setIsRegister] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);
  
  const { login, register, isLoading, error, clearError } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError(null);
    clearError();

    if (!email || !password) {
      setLocalError('Please fill in all fields');
      return;
    }

    try {
      if (isRegister) {
        await register(email, password);
      } else {
        await login(email, password);
      }
      setView('home');
    } catch {
      // Error is already set in auth context
    }
  };

  const displayError = localError || error;

  return (
    <div className="max-w-md mx-auto animate-in fade-in slide-in-from-bottom-8 duration-500 pt-12 pb-20">
      <div className="bg-white border-2 border-zinc-900 p-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] relative">
        <div className="absolute top-0 right-0 p-2 bg-black text-white text-[10px] font-mono font-bold uppercase">
          {isRegister ? 'REGISTER' : 'AUTH_V2'}
        </div>

        <div className="mb-8 text-center">
          <div className="w-12 h-12 bg-[#E11D48] mx-auto flex items-center justify-center mb-4">
            <Lock className="w-6 h-6 text-white" />
          </div>
          <h2 className="text-2xl font-bold tracking-tight text-zinc-900 uppercase">
            {isRegister ? 'Create Account' : 'System Access'}
          </h2>
          <p className="text-zinc-500 text-sm mt-2 font-mono">
            {isRegister ? 'Sign up for a new account' : 'Enter credentials to continue'}
          </p>
        </div>

        {displayError && (
          <div className="mb-6 p-3 bg-red-50 border-2 border-red-200 text-red-700 text-sm font-mono">
            {displayError}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <label className="text-xs font-bold uppercase tracking-wider flex items-center gap-2 text-zinc-700">
              <Mail className="w-3 h-3" />
              Email Address
            </label>
            <input 
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="user@example.com"
              className="w-full bg-zinc-50 border-2 border-zinc-200 p-3 text-sm font-mono focus:border-black focus:ring-0 outline-none transition-colors placeholder:text-zinc-400"
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <label className="text-xs font-bold uppercase tracking-wider flex items-center gap-2 text-zinc-700">
              <ShieldCheck className="w-3 h-3" />
              Password
            </label>
            <div className="relative">
              <input 
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••••••"
                className="w-full bg-zinc-50 border-2 border-zinc-200 p-3 text-sm font-mono focus:border-black focus:ring-0 outline-none transition-colors placeholder:text-zinc-400"
                disabled={isLoading}
              />
              <button 
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-black transition-colors"
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
            {isRegister && (
              <p className="text-[10px] text-zinc-400 font-mono">Minimum 8 characters required</p>
            )}
          </div>

          <button 
            type="submit"
            disabled={isLoading}
            className="w-full bg-black text-white py-3 font-bold uppercase tracking-wider hover:bg-[#E11D48] transition-colors border border-black flex items-center justify-center gap-2 group disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Processing...' : (isRegister ? 'Create Account' : 'Sign In')}
            <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
          </button>
        </form>

        <div className="my-8 flex items-center gap-4">
          <div className="h-px bg-zinc-200 flex-1"></div>
          <span className="text-[10px] font-bold uppercase text-zinc-400 tracking-widest">Or continue with</span>
          <div className="h-px bg-zinc-200 flex-1"></div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <button className="flex items-center justify-center gap-2 p-3 border-2 border-zinc-200 hover:border-black hover:bg-zinc-50 transition-all group opacity-50 cursor-not-allowed">
            <Github className="w-5 h-5 group-hover:scale-110 transition-transform" />
            <span className="text-xs font-bold uppercase">Github</span>
          </button>
          <button className="flex items-center justify-center gap-2 p-3 border-2 border-zinc-200 hover:border-black hover:bg-zinc-50 transition-all group opacity-50 cursor-not-allowed">
            <GoogleIcon />
            <span className="text-xs font-bold uppercase">Google</span>
          </button>
        </div>

        <div className="mt-8 text-center space-y-2">
          <button 
            onClick={() => {
              setIsRegister(!isRegister);
              setLocalError(null);
              clearError();
            }}
            className="text-xs font-mono text-[#E11D48] hover:text-black underline decoration-1 underline-offset-4"
          >
            {isRegister ? 'Already have an account? Sign in' : "Don't have an account? Register"}
          </button>
          {!isRegister && (
            <div>
              <button 
                onClick={() => setView('forgot-password')} 
                className="text-xs font-mono text-zinc-500 hover:text-[#E11D48] underline decoration-1 underline-offset-4"
              >
                Forgot your password?
              </button>
            </div>
          )}
        </div>
      </div>
      
      <div className="mt-6 text-center text-xs text-zinc-400 font-mono">SECURE CONNECTION // 256-BIT ENCRYPTION</div>
    </div>
  );
}
