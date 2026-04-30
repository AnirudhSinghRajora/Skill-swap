'use client';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useState, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { Check, Circle } from 'lucide-react';
import api, { ApiClientError } from '@/lib/api';
import { completePasswordAuthWithE2EE } from '@/lib/e2ee/auth';

interface UserDataType {
  name: string;
  email: string;
  password: string;
}

export function SignupForm({ className, ...props }: React.ComponentProps<'div'>) {
  const [userData, setUserData] = useState<UserDataType>({ name: '', email: '', password: '' });
  const [error, setError] = useState('');
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const passwordChecks = useMemo(() => [
    { label: 'At least 8 characters', met: userData.password.length >= 8 },
    { label: 'Contains a letter', met: /[a-zA-Z]/.test(userData.password) },
    { label: 'Contains a number', met: /[0-9]/.test(userData.password) },
    { label: 'Contains a special character', met: /[^a-zA-Z0-9]/.test(userData.password) },
  ], [userData.password]);

  const allChecksMet = passwordChecks.every((check) => check.met);

  const handleSignup = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError('');

    if (!allChecksMet) {
      setError('Please meet all password requirements');
      return;
    }

    setLoading(true);

    try {
      const data = await api.auth.register({
        name: userData.name,
        email: userData.email,
        password: userData.password,
      });

      try {
        await completePasswordAuthWithE2EE(data, userData.password);
      } catch (e) {
        console.error('E2EE setup failed after signup', e);
        setError('Account created, but secure messaging setup failed. Please sign in again to retry.');
        return;
      }

      router.push('/profile');
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

  return (
    <div className={cn('w-full max-w-sm mx-auto', className)} {...props}>
      <div className="mb-8 text-center">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Create an account</h1>
        <p className="text-sm text-muted-foreground mt-1.5">Get started with SkillSwap</p>
      </div>

      <form onSubmit={handleSignup} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="name" className="text-sm">Full Name</Label>
          <Input id="name" required placeholder="Jane Doe"
            value={userData.name}
            onChange={(e) => setUserData({ ...userData, name: e.target.value })}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="email" className="text-sm">Email</Label>
          <Input id="email" type="email" required placeholder="you@example.com"
            value={userData.email}
            onChange={(e) => setUserData({ ...userData, email: e.target.value })}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="password" className="text-sm">Password</Label>
          <Input id="password" type="password" required
            value={userData.password}
            onChange={(e) => setUserData({ ...userData, password: e.target.value })}
            placeholder="Create a strong password"
          />
          {userData.password.length > 0 && (
            <ul className="mt-2 space-y-1">
              {passwordChecks.map((check) => (
                <li key={check.label} className="flex items-center gap-2 text-xs">
                  {check.met ? (
                    <Check className="h-3 w-3 text-success shrink-0" />
                  ) : (
                    <Circle className="h-3 w-3 text-muted-foreground/40 shrink-0" />
                  )}
                  <span className={check.met ? 'text-success' : 'text-muted-foreground'}>
                    {check.label}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>

        {error && <p className="text-destructive text-sm">{error}</p>}

        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? 'Creating account…' : 'Create account'}
        </Button>
      </form>

      <p className="text-center text-sm text-muted-foreground mt-6">
        Already have an account?{' '}
        <a href="/auth/signin" className="text-primary hover:underline font-medium">Sign in</a>
      </p>

      <p className="text-center text-xs text-muted-foreground mt-8">
        By continuing, you agree to our <a href="#" className="underline underline-offset-2">Terms</a> &{' '}
        <a href="#" className="underline underline-offset-2">Privacy Policy</a>.
      </p>
    </div>
  );
}
