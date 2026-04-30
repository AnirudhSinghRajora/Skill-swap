'use client';

import { useEffect, useState, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { CheckCircle2, XCircle, Loader2, MailWarning } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import api from '@/lib/api';

type Status = 'pending' | 'success' | 'error' | 'missing';

function VerifyEmailInner() {
  const params = useSearchParams();
  const token = params.get('token');
  const [status, setStatus] = useState<Status>(token ? 'pending' : 'missing');
  const [message, setMessage] = useState<string>('');

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    (async () => {
      try {
        const res = await api.auth.verifyEmail(token);
        if (cancelled) return;
        setStatus('success');
        setMessage(res.message ?? 'Your email has been verified.');
      } catch (err) {
        if (cancelled) return;
        setStatus('error');
        setMessage(err instanceof Error ? err.message : 'Verification failed.');
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [token]);

  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-muted">
            {status === 'pending' && <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />}
            {status === 'success' && <CheckCircle2 className="h-6 w-6 text-emerald-600" />}
            {status === 'error' && <XCircle className="h-6 w-6 text-destructive" />}
            {status === 'missing' && <MailWarning className="h-6 w-6 text-amber-600" />}
          </div>
          <CardTitle>
            {status === 'pending' && 'Verifying your email…'}
            {status === 'success' && 'Email verified'}
            {status === 'error' && 'Verification failed'}
            {status === 'missing' && 'Missing verification token'}
          </CardTitle>
          <CardDescription>
            {status === 'pending' && 'Hang tight while we confirm your address.'}
            {status === 'success' && message}
            {status === 'error' && message}
            {status === 'missing' &&
              'This link is incomplete. Open the most recent email from us and click the button again.'}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          {status === 'success' && (
            <Button asChild>
              <Link href="/dashboard">Go to dashboard</Link>
            </Button>
          )}
          {(status === 'error' || status === 'missing') && (
            <>
              <Button asChild variant="outline">
                <Link href="/dashboard">Go to dashboard</Link>
              </Button>
              <p className="text-center text-xs text-muted-foreground">
                You can request a new link from your dashboard.
              </p>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-background flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <VerifyEmailInner />
    </Suspense>
  );
}
