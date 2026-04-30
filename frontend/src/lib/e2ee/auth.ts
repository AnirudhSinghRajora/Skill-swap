import {
  auth as authApi,
  clearAuth,
  notifyAuthChange,
  setTokens,
  type AuthResponse,
} from '@/lib/api';
import { initializeE2EE } from '@/lib/e2ee/init';

export class E2EEOnboardingError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'E2EEOnboardingError';
  }
}


export async function completePasswordAuthWithE2EE(
  authResponse: AuthResponse,
  password: string,
): Promise<void> {
  setTokens(authResponse.access_token, authResponse.refresh_token);

  try {
    await initializeE2EE(password, authResponse.user.user_id);

    const me = await authApi.me();
    if (!me.has_public_key || !me.has_key_backup) {
      throw new E2EEOnboardingError(
        'Secure messaging setup did not complete. Please sign in again to retry.',
      );
    }

    localStorage.setItem('user', JSON.stringify(authResponse.user));
    notifyAuthChange();
  } catch (error) {
    clearAuth();
    throw error;
  }
}