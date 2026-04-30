'use client';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { SiGoogle } from 'react-icons/si';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import api, { ApiClientError } from '@/lib/api';
import { E2EERestoreError } from '@/lib/e2ee/init';
import { completePasswordAuthWithE2EE } from '@/lib/e2ee/auth';

interface UserDataType {
  email: string;
  password: string;
}

export function LoginForm({ className, ...props }: React.ComponentProps<'div'>) {
  const [userData, setUserData] = useState<UserDataType>({ email: '', password: '' });
  const [error, setError] = useState('');
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const handleLogin = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError('');
    setLoading(true);

    try {
      const data = await api.auth.login(userData.email, userData.password);

      try {
        await completePasswordAuthWithE2EE(data, userData.password);
      } catch (e) {
        const msg =
          e instanceof E2EERestoreError
            ? e.message
            : 'Signed in, but secure messaging setup failed. Please try again.';
        setError(msg);
        return;
      }

      router.push('/dashboard');
    } catch (err: unknown) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError('An error occurred');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleSocialLogin = (provider: string) => {
    if (provider === 'google') {
      window.location.href = `${process.env.NEXT_PUBLIC_API_BASE_URL}/auth/google`;
    }
  };

  return (
    <div className={cn('w-full max-w-sm mx-auto', className)} {...props}>
      <div className="mb-8 text-center">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Welcome back</h1>
        <p className="text-sm text-muted-foreground mt-1.5">Sign in to your account</p>
      </div>

      <form onSubmit={handleLogin} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="email" className="text-sm">Email</Label>
          <Input id="email" type="email" required placeholder="you@example.com"
            value={userData.email}
            onChange={(e) => setUserData({ ...userData, email: e.target.value })}
          />
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label htmlFor="password" className="text-sm">Password</Label>
            <a href="/auth/forgot-password" className="text-xs text-muted-foreground hover:text-foreground transition-colors">
              Forgot password?
            </a>
          </div>
          <Input id="password" type="password" required
            value={userData.password}
            onChange={(e) => setUserData({ ...userData, password: e.target.value })}
            placeholder="Enter your password"
          />
        </div>

        {error && <p className="text-destructive text-sm">{error}</p>}

        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? 'Signing in…' : 'Sign in'}
        </Button>
      </form>

      <div className="relative my-6">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-border" />
        </div>
        <div className="relative flex justify-center text-xs">
          <span className="bg-background px-2 text-muted-foreground">or</span>
        </div>
      </div>

      <Button variant="outline" type="button" className="w-full"
        disabled={loading}
        onClick={() => handleSocialLogin('google')}
      >
        <SiGoogle className="h-4 w-4 mr-2" /> Continue with Google
      </Button>

      <p className="text-center text-sm text-muted-foreground mt-6">
        Don&apos;t have an account?{' '}
        <a href="/auth/signup" className="text-primary hover:underline font-medium">Sign up</a>
      </p>

      <p className="text-center text-xs text-muted-foreground mt-8">
        By continuing, you agree to our <a href="#" className="underline underline-offset-2">Terms</a> &{' '}
        <a href="#" className="underline underline-offset-2">Privacy Policy</a>.
      </p>
    </div>
  );
}
