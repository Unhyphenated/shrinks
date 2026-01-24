import { useState } from "react";
import type { ClicksByDate } from "../types";

interface BarChartProps {
  data?: ClicksByDate[];
  isLoading?: boolean;
}

export function BarChart({ data, isLoading }: BarChartProps) {
  // Generate stable random heights only once for skeleton loaders
  const [skeletonHeights] = useState(() =>
    Array.from({ length: 14 }, () => 30 + Math.random() * 50),
  );

  // Use provided data or fallback to placeholder
  const chartData =
    data && data.length > 0
      ? data.map((d) => d.clicks)
      : [45, 70, 35, 60, 80, 50, 90, 65, 40, 55, 75, 60, 85, 95];

  const maxValue = Math.max(...chartData, 1);

  if (isLoading) {
    return (
      <div className="h-64 flex items-end justify-between gap-1 pt-8 pb-2 px-2 border-b-2 border-black">
        {skeletonHeights.map((height, i) => (
          <div
            key={i}
            className="group relative flex-1 flex flex-col justify-end h-full"
          >
            <div
              className="w-full bg-zinc-100 animate-pulse"
              style={{ height: `${height}%` }}
            />
            <div className="h-2 w-full mt-2 border-t border-zinc-300"></div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="h-64 flex items-end justify-between gap-1 pt-8 pb-2 px-2 border-b-2 border-black">
      {chartData.map((value, i) => {
        const height = (value / maxValue) * 100;
        const label = data?.[i]?.date || `Day ${i + 1}`;
        return (
          <div
            key={i}
            className="group relative flex-1 flex flex-col justify-end h-full hover:bg-zinc-50 transition-colors"
          >
            <div
              className="w-full bg-zinc-200 group-hover:bg-[#E11D48] transition-all relative"
              style={{ height: `${height}%` }}
            >
              <div className="absolute -top-8 left-1/2 -translate-x-1/2 bg-black text-white text-[10px] font-mono py-1 px-2 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-10">
                {value} clicks
                <br />
                <span className="text-zinc-400">{label}</span>
              </div>
            </div>
            <div className="h-2 w-full mt-2 border-t border-zinc-300"></div>
          </div>
        );
      })}
    </div>
  );
}
