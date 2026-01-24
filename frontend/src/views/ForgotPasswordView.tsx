import { useState } from "react";
import { Key, Mail, ArrowRight, ArrowLeft } from "lucide-react";
import type { ViewState } from "../types";

interface ForgotPasswordViewProps {
  setView: (v: ViewState) => void;
}

export function ForgotPasswordView({ setView }: ForgotPasswordViewProps) {
  const [email, setEmail] = useState("");
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // TODO: Wire up to backend when password reset endpoint is available
    setIsSubmitted(true);
  };

  return (
    <div className="max-w-md mx-auto animate-in fade-in slide-in-from-bottom-8 duration-500 pt-12 pb-20">
      <div className="bg-white border-2 border-zinc-900 p-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] relative">
        <div className="absolute top-0 right-0 p-2 bg-black text-white text-[10px] font-mono font-bold uppercase">
          RECOVERY
        </div>

        <div className="mb-8 text-center">
          <div className="w-12 h-12 bg-zinc-100 mx-auto flex items-center justify-center mb-4 border-2 border-zinc-900">
            <Key className="w-6 h-6 text-[#E11D48]" />
          </div>
          <h2 className="text-2xl font-bold tracking-tight text-zinc-900 uppercase">
            Reset Password
          </h2>
          <p className="text-zinc-500 text-sm mt-2 font-mono">
            {isSubmitted
              ? "Check your email for instructions."
              : "Enter your email to receive recovery instructions."}
          </p>
        </div>

        {isSubmitted ? (
          <div className="text-center">
            <div className="mb-6 p-4 bg-green-50 border-2 border-green-200 text-green-700 text-sm font-mono">
              If an account exists with that email, you'll receive a password
              reset link shortly.
            </div>
            <button
              onClick={() => setView("login")}
              className="text-xs font-bold uppercase tracking-wider text-zinc-500 hover:text-black flex items-center justify-center gap-2 mx-auto group"
            >
              <ArrowLeft className="w-3 h-3 group-hover:-translate-x-1 transition-transform" />
              Return to Sign In
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label className="text-xs font-bold uppercase tracking-wider flex items-center gap-2 text-zinc-700">
                <Mail className="w-3 h-3" />
                Registered Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="user@example.com"
                className="w-full bg-zinc-50 border-2 border-zinc-200 p-3 text-sm font-mono focus:border-black focus:ring-0 outline-none transition-colors placeholder:text-zinc-400"
                required
              />
            </div>

            <button
              type="submit"
              className="w-full bg-[#E11D48] text-white py-3 font-bold uppercase tracking-wider hover:bg-black transition-colors border border-black flex items-center justify-center gap-2 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[2px] active:translate-y-[2px]"
            >
              Send Reset Link
              <ArrowRight className="w-4 h-4" />
            </button>
          </form>
        )}

        {!isSubmitted && (
          <div className="mt-8 pt-6 border-t border-zinc-100 text-center">
            <button
              onClick={() => setView("login")}
              className="text-xs font-bold uppercase tracking-wider text-zinc-500 hover:text-black flex items-center justify-center gap-2 mx-auto group"
            >
              <ArrowLeft className="w-3 h-3 group-hover:-translate-x-1 transition-transform" />
              Return to Sign In
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
