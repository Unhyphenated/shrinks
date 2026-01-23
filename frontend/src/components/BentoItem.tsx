import type { LucideIcon } from 'lucide-react';

interface BentoItemProps {
  title: string;
  value: string;
  icon: LucideIcon;
  colSpan?: string;
}

export function BentoItem({ title, value, icon: Icon, colSpan = "col-span-1" }: BentoItemProps) {
  return (
    <div className={`${colSpan} bg-white border border-zinc-200 p-6 relative overflow-hidden group hover:border-black transition-colors`}>
      <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
        <Icon className="w-20 h-20 text-black" />
      </div>
      <div className="relative z-10">
        <div className="flex items-center gap-2 mb-4 text-zinc-500">
          <Icon className="w-4 h-4" />
          <span className="text-xs font-mono uppercase tracking-wider font-bold">{title}</span>
        </div>
        <div className="flex items-center gap-3">
          <div className="w-1 h-12 bg-[#E11D48]"></div>
          <div className="text-4xl font-bold text-zinc-900 tracking-tight font-mono">{value}</div>
        </div>
      </div>
    </div>
  );
}
