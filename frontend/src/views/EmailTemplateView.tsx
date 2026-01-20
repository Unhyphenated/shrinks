import { Copy } from 'lucide-react';

export function EmailTemplateView() {
  const handleCopyHtml = () => {
    // Copy the email HTML template
    const emailHtml = document.querySelector('.email-template-container')?.innerHTML || '';
    navigator.clipboard.writeText(emailHtml);
  };

  return (
    <div className="max-w-2xl mx-auto animate-in fade-in slide-in-from-bottom-8 duration-500 pt-8 pb-20">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h2 className="text-3xl font-bold text-zinc-900 tracking-tighter uppercase">Template Preview</h2>
          <p className="text-zinc-500 text-sm font-mono">Subject: Reset your Shrinks Password</p>
        </div>
        <button 
          onClick={handleCopyHtml}
          className="px-4 py-2 border-2 border-zinc-200 hover:border-black bg-white text-xs font-bold uppercase tracking-wider transition-colors flex items-center gap-2"
        >
          <Copy className="w-3 h-3" /> Copy HTML
        </button>
      </div>

      {/* Email Container Imitation */}
      <div className="bg-zinc-100 p-8 rounded-lg border border-zinc-200">
        <div className="email-template-container bg-white max-w-lg mx-auto border border-zinc-200 shadow-sm">
          
          {/* Email Header */}
          <div className="bg-zinc-900 p-6 text-center">
            <span className="text-lg font-bold tracking-tight text-white">
              Shrinks<span className="text-[#E11D48]">.</span>
            </span>
          </div>

          {/* Email Body */}
          <div className="p-8">
            <h1 className="text-xl font-bold text-zinc-900 mb-4 font-sans">Password Reset Request</h1>
            
            <p className="text-zinc-600 text-sm leading-relaxed mb-6 font-sans">
              Hello,
              <br/><br/>
              We received a request to reset the password for your Shrinks account associated with this email address. If you didn't make this request, you can safely ignore this email.
            </p>

            {/* CTA Button */}
            <div className="text-center my-8">
              <a 
                href="#" 
                className="inline-block bg-[#E11D48] text-white px-8 py-3 text-sm font-bold uppercase tracking-wider border border-black hover:bg-black transition-colors"
                style={{ textDecoration: 'none' }}
              >
                Reset Password
              </a>
            </div>

            <p className="text-zinc-600 text-sm leading-relaxed mb-6 font-sans">
              This link will expire in 60 minutes. If you need assistance, please contact our support team.
              <br/><br/>
              — The Shrinks Team
            </p>
            
            <div className="border-t border-zinc-100 pt-6 mt-6">
              <p className="text-xs text-zinc-400 font-mono">
                If you're having trouble clicking the "Reset Password" button, copy and paste the URL below into your web browser:
              </p>
              <p className="text-xs text-[#E11D48] font-mono mt-2 break-all underline">
                https://shrinks.io/auth/reset-password/8392-1293-8493-1029
              </p>
            </div>
          </div>

          {/* Email Footer */}
          <div className="bg-zinc-50 p-6 text-center border-t border-zinc-200">
            <p className="text-[10px] text-zinc-400 font-sans uppercase tracking-wider">
              © 2024 Shrinks Inc. All rights reserved.
            </p>
          </div>

        </div>
      </div>
    </div>
  );
}
