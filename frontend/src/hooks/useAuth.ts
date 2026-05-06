'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { getAccessToken, clearAuth, decodeTokenPayload, type UserInfo } from '@/lib/api';

interface AuthState {
  user: UserInfo | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

function loadAuthState(): AuthState {
  const token = getAccessToken();
  if (!token) {
    return { user: null, isAuthenticated: false, isLoading: false };
  }

  // Derive identity from the signed JWT — this can't be tampered with.
  // localStorage 'user' is only a display cache (name, photo, etc.).
  const claims = decodeTokenPayload(token);
  if (!claims) {
    return { user: null, isAuthenticated: false, isLoading: false };
  }

  let displayCache: Partial<UserInfo> = {};
  try {
    const stored = localStorage.getItem('user');
    if (stored) displayCache = JSON.parse(stored);
  } catch { /* ignore */ }

  const user: UserInfo = {
    user_id: claims.user_id,
    email: claims.email,
    name: displayCache.name || claims.email,
    location: displayCache.location ?? null,
    has_photo: displayCache.has_photo ?? false,
    is_public: displayCache.is_public ?? true,
  };

  return { user, isAuthenticated: true, isLoading: false };
}

export function useAuth(requireAuth: boolean = false) {
  const [state, setState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true,
  });
  const router = useRouter();

  useEffect(() => {
    const authState = loadAuthState();
    setState(authState);
    if (!authState.isAuthenticated && requireAuth) {
      router.replace('/auth/signin');
    }
  }, [requireAuth, router]);

  // Listen for auth changes from other components
  useEffect(() => {
    const handleAuthChange = () => {
      const authState = loadAuthState();
      setState(authState);
      if (!authState.isAuthenticated && requireAuth) {
        router.replace('/auth/signin');
      }
    };
    window.addEventListener('auth-change', handleAuthChange);
    return () => window.removeEventListener('auth-change', handleAuthChange);
  }, [requireAuth, router]);

  const logout = useCallback(() => {
    clearAuth();
    setState({ user: null, isAuthenticated: false, isLoading: false });
    router.push('/auth/signin');
  }, [router]);

  return { ...state, logout };
}
