'use client';

import { useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { setTokens, notifyAuthChange, decodeTokenPayload } from '@/lib/api';

/**
 * Google OAuth callback.
 *
 * NOTE on E2EE: OAuth users have no plaintext password client-side, so we
 * cannot derive a PBKDF2 key here for the encrypted private-key backup.
 * E2EE is therefore disabled for OAuth-only accounts on this device until
 * the user sets a local password (future work). We deliberately do NOT
 * call initializeE2EE() here — it would silently rotate keys.
 */
function CallbackHandler() {
	const router = useRouter();
	const searchParams = useSearchParams();

	useEffect(() => {
		const accessToken = searchParams.get('access_token');
		const refreshToken = searchParams.get('refresh_token');
		const error = searchParams.get('error');

		if (error) {
			router.replace(`/auth/signin?error=${encodeURIComponent(error)}`);
			return;
		}

		if (!accessToken || !refreshToken) {
			router.replace('/auth/signin?error=missing_tokens');
			return;
		}

		setTokens(accessToken, refreshToken);

		// Seed the display cache from JWT claims so useAuth has a name to show
		// before any /users/me round-trip.
		const claims = decodeTokenPayload(accessToken);
		if (claims) {
			try {
				localStorage.setItem(
					'user',
					JSON.stringify({
						user_id: claims.user_id,
						email: claims.email,
						name: claims.email,
					}),
				);
			} catch {
				/* private-mode storage failures — non-fatal */
			}
		}

		notifyAuthChange();
		router.replace('/dashboard');
	}, [searchParams, router]);

	return (
		<div className="flex min-h-svh items-center justify-center bg-background">
			<div className="text-center">
				<div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto mb-4" />
				<p className="text-sm text-muted-foreground">Signing you in…</p>
			</div>
		</div>
	);
}

export default function AuthCallbackPage() {
	return (
		<Suspense
			fallback={
				<div className="flex min-h-svh items-center justify-center bg-background">
					<div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
				</div>
			}
		>
			<CallbackHandler />
		</Suspense>
	);
}
