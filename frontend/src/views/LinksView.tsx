import { useState, useEffect } from "react";
import {
  Copy,
  ChevronLeft,
  ChevronRight,
  Filter,
  Link as LinkIcon,
  Trash2,
  BarChart2,
  ArrowLeft,
  Check,
} from "lucide-react";
import { apiClient } from "../api/client";
import type { Link, ViewState } from "../types";

// Placeholder domain - replace when you secure a domain
const SHORT_DOMAIN = "shrinks.io";

interface LinksViewProps {
  setView: (v: ViewState) => void;
  setSelectedLinkCode: (code: string) => void;
}

export function LinksView({ setView, setSelectedLinkCode }: LinksViewProps) {
  const [links, setLinks] = useState<Link[]>([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [isLoading, setIsLoading] = useState(true);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [deleteConfirm, setDeleteConfirm] = useState<{
    shortCode: string;
    linkId: number;
  } | null>(null);

  const itemsPerPage = 10;

  const fetchLinks = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const offset = (currentPage - 1) * itemsPerPage;
      const response = await apiClient.getLinks(itemsPerPage, offset);
      setLinks(response.links || []);
      setTotal(response.total || 0);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch links";
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchLinks();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPage]);

  const handleDeleteClick = (shortCode: string, linkId: number) => {
    setDeleteConfirm({ shortCode, linkId });
  };

  const handleDeleteConfirm = async () => {
    if (!deleteConfirm) return;

    setDeletingId(deleteConfirm.linkId);
    try {
      await apiClient.deleteLink(deleteConfirm.shortCode);
      setDeleteConfirm(null);
      // Refresh the list
      await fetchLinks();
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to delete link";
      setError(message);
    } finally {
      setDeletingId(null);
    }
  };

  const handleViewAnalytics = (shortCode: string) => {
    setSelectedLinkCode(shortCode);
    setView("analytics");
  };

  const handleCopy = (shortCode: string) => {
    navigator.clipboard.writeText(`${SHORT_DOMAIN}/${shortCode}`);

    setCopiedId(shortCode);

    setTimeout(() => setCopiedId(null), 2000);
  };
  const totalPages = Math.ceil(total / itemsPerPage);

  // Filter links by search query
  const filteredLinks = searchQuery
    ? links.filter(
        (link) =>
          link.short_code.toLowerCase().includes(searchQuery.toLowerCase()) ||
          link.long_url.toLowerCase().includes(searchQuery.toLowerCase()),
      )
    : links;

  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
      {/* Back Button */}
      <button
        onClick={() => setView("home")}
        className="inline-flex items-center gap-2 px-4 py-2 mb-6 border-2 border-zinc-200 bg-white hover:border-black hover:cursor-pointer transition-colors text-sm font-mono uppercase font-bold"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Home
      </button>

      <div className="flex items-center justify-between mb-12">
        <div>
          <div className="inline-flex items-center gap-2 px-3 py-1 bg-zinc-100 border border-zinc-200 text-zinc-900 text-xs font-mono mb-4 font-bold uppercase tracking-wider">
            <LinkIcon className="w-3 h-3" />
            <span>Link Manager</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-zinc-900 tracking-tighter uppercase">
            Your <span className="text-[#E11D48]">Links</span>
          </h2>
        </div>

        <div className="flex items-center gap-2">
          <div className="relative hidden md:block">
            <input
              type="text"
              placeholder="Search links..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10 pr-4 py-2 border-2 border-zinc-200 focus:border-black outline-none font-mono text-sm w-64 bg-white placeholder-zinc-400 transition-colors"
            />
            <Filter className="w-4 h-4 text-zinc-400 absolute left-3 top-1/2 -translate-y-1/2" />
          </div>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border-2 border-red-200 text-red-700 font-mono text-sm">
          {error}
        </div>
      )}

      {/* Table Container */}
      <div className="border-2 border-zinc-200 bg-white shadow-[8px_8px_0px_0px_rgba(0,0,0,0.1)]">
        <div className="grid grid-cols-12 gap-4 p-4 border-b-2 border-zinc-200 bg-zinc-50 text-xs font-bold uppercase tracking-wider text-zinc-500 font-mono">
          <div className="col-span-5 md:col-span-4">URL Details</div>
          <div className="col-span-3 md:col-span-2 text-right md:text-left">
            Short Code
          </div>
          <div className="hidden md:block col-span-2">Created</div>
          <div className="hidden md:block col-span-2">Status</div>
          <div className="col-span-4 md:col-span-2 text-right">Actions</div>
        </div>

        {isLoading ? (
          <div className="divide-y divide-zinc-100">
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="grid grid-cols-12 gap-4 p-4 items-center animate-pulse"
              >
                <div className="col-span-5 md:col-span-4">
                  <div className="h-4 bg-zinc-200 rounded w-3/4 mb-2"></div>
                  <div className="h-3 bg-zinc-100 rounded w-1/2"></div>
                </div>
                <div className="col-span-3 md:col-span-2">
                  <div className="h-4 bg-zinc-200 rounded w-20"></div>
                </div>
                <div className="hidden md:block col-span-2">
                  <div className="h-4 bg-zinc-100 rounded w-24"></div>
                </div>
                <div className="hidden md:block col-span-2">
                  <div className="h-4 bg-zinc-100 rounded w-16"></div>
                </div>
                <div className="col-span-4 md:col-span-2">
                  <div className="h-8 bg-zinc-100 rounded w-full"></div>
                </div>
              </div>
            ))}
          </div>
        ) : filteredLinks.length === 0 ? (
          <div className="p-12 text-center">
            <p className="text-zinc-500 font-mono">
              {searchQuery
                ? "No links match your search."
                : "No links yet. Create your first short link!"}
            </p>
          </div>
        ) : (
          <div className="divide-y divide-zinc-100">
            {filteredLinks.map((link) => (
              <div
                key={link.id}
                className="grid grid-cols-12 gap-4 p-4 items-center hover:bg-zinc-50 group transition-colors"
              >
                <div className="col-span-5 md:col-span-4 overflow-hidden">
                  <div className="font-bold text-zinc-900 font-mono truncate">
                    {SHORT_DOMAIN}/{link.short_code}
                  </div>
                  <div className="text-xs text-zinc-400 truncate mt-1 group-hover:text-[#E11D48] transition-colors">
                    {link.long_url}
                  </div>
                </div>

                <div className="col-span-3 md:col-span-2 text-right md:text-left font-mono text-sm text-zinc-600">
                  {link.short_code}
                </div>

                <div className="hidden md:block col-span-2 text-xs text-zinc-500 font-mono">
                  {new Date(link.created_at).toLocaleDateString()}
                </div>

                <div className="hidden md:block col-span-2">
                  <span className="inline-flex items-center px-2 py-1 border border-green-200 bg-green-50 text-green-700 text-[10px] font-bold uppercase tracking-wider">
                    Active
                  </span>
                </div>

                <div className="col-span-4 md:col-span-2 flex justify-end items-center gap-1">
                  <button
                    onClick={() => handleCopy(link.short_code)}
                    className="p-2 hover:bg-zinc-200 hover:cursor-pointer transition-colors"
                    title="Copy"
                  >
                    {copiedId === link.short_code ? (
                      <Check className="w-4 h-4 text-green-600" />
                    ) : (
                      <Copy className="w-4 h-4 text-zinc-400 hover:text-black" />
                    )}
                  </button>
                  <button
                    onClick={() => handleViewAnalytics(link.short_code)}
                    className="p-2 hover:bg-zinc-200 hover:cursor-pointer transition-colors"
                    title="View Analytics"
                  >
                    <BarChart2 className="w-4 h-4 text-zinc-400 hover:text-black" />
                  </button>
                  <button
                    onClick={() => handleDeleteClick(link.short_code, link.id)}
                    disabled={deletingId === link.id}
                    className="p-2 hover:bg-red-100 hover:cursor-pointer transition-colors disabled:opacity-50"
                    title="Delete"
                  >
                    <Trash2 className="w-4 h-4 text-zinc-400 hover:text-red-600" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination Footer */}
        {!isLoading && total > 0 && (
          <div className="p-4 border-t-2 border-zinc-200 bg-zinc-50 flex items-center justify-between">
            <div className="text-xs font-mono text-zinc-500 uppercase">
              Showing{" "}
              <span className="text-black font-bold">
                {(currentPage - 1) * itemsPerPage + 1}
              </span>{" "}
              -{" "}
              <span className="text-black font-bold">
                {Math.min(currentPage * itemsPerPage, total)}
              </span>{" "}
              of {total}
            </div>

            <div className="flex gap-2">
              <button
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                disabled={currentPage === 1}
                className="p-2 border-2 border-zinc-200 bg-white hover:cursor-pointer hover:border-black disabled:opacity-50 disabled:hover:border-zinc-200 transition-colors"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>

              {Array.from({ length: Math.min(5, totalPages) }).map((_, i) => {
                const pageNum = i + 1;
                return (
                  <button
                    key={i}
                    onClick={() => setCurrentPage(pageNum)}
                    className={`w-9 h-9 border-2 flex items-center hover:cursor-pointer justify-center text-sm font-bold font-mono transition-colors ${
                      currentPage === pageNum
                        ? "border-black bg-black text-white"
                        : "border-zinc-200 bg-white hover:border-black text-zinc-600"
                    }`}
                  >
                    {pageNum}
                  </button>
                );
              })}

              <button
                onClick={() =>
                  setCurrentPage((p) => Math.min(totalPages, p + 1))
                }
                disabled={currentPage === totalPages || totalPages === 0}
                className="p-2 border-2 border-zinc-200 bg-white hover:cursor-pointer hover:border-black disabled:opacity-50 disabled:hover:border-zinc-200 transition-colors"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Custom Delete Confirmation Dialog */}
      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white border-4 border-black shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] max-w-md w-full p-8">
            <h3 className="text-2xl font-bold text-zinc-900 mb-4 uppercase tracking-tighter">
              Delete Link?
            </h3>
            <p className="text-zinc-600 mb-2 font-mono text-sm">
              Are you sure you want to delete this link?
            </p>
            <p className="text-sm text-zinc-500 mb-6 font-mono break-all">
              <span className="font-bold text-[#E11D48]">
                {SHORT_DOMAIN}/{deleteConfirm.shortCode}
              </span>
            </p>
            <p className="text-xs text-zinc-400 mb-6 font-mono">
              This action cannot be undone. All analytics data for this link
              will also be deleted.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setDeleteConfirm(null)}
                className="flex-1 px-4 py-3 border-2 border-zinc-200 bg-white hover:bg-zinc-50 hover:cursor-pointer font-bold text-sm uppercase transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteConfirm}
                disabled={deletingId !== null}
                className="flex-1 px-4 py-3 border-2 border-red-600 bg-red-600 text-white hover:bg-red-700 hover:cursor-pointer hover:border-red-700 font-bold text-sm uppercase transition-colors disabled:opacity-50"
              >
                {deletingId === deleteConfirm.linkId ? "Deleting..." : "Delete"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
