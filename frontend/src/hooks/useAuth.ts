import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  createElement,
} from "react";
import type { ReactNode } from "react";
import {
  apiClient,
  clearTokens,
  setTokens,
  getAccessToken,
} from "../api/client";
import type { User } from "../types";

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const isAuthenticated = !!user && !!getAccessToken();

  // Set up unauthorized callback
  // 1. New Effect to check for existing session on mount
  useEffect(() => {
    async function initAuth() {
      const token = getAccessToken();

      // If no token, we aren't logged in, just stop loading
      if (!token) {
        setIsLoading(false);
        return;
      }

      try {
        // 2. Call your new Go endpoint
        const userData = await apiClient.getCurrentUser();
        setUser(userData);
      } catch (err) {
        // If the token is expired or invalid, wipe it
        console.error("Session restoration failed:", err);
        clearTokens();
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    }

    initAuth();
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiClient.login(email, password);
      setUser(response.user);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Login failed";
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const register = useCallback(
    async (email: string, password: string) => {
      setIsLoading(true);
      setError(null);
      try {
        await apiClient.register(email, password);
        // Auto-login after registration
        await login(email, password);
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Registration failed";
        setError(message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [login],
  );

  const logout = useCallback(async () => {
    setIsLoading(true);
    try {
      await apiClient.logout();
    } finally {
      setUser(null);
      clearTokens();
      setIsLoading(false);
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const value: AuthContextType = {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    register,
    logout,
    clearError,
  };

  return createElement(AuthContext.Provider, { value }, children);
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

// Re-export for convenience
export { setTokens, clearTokens };
