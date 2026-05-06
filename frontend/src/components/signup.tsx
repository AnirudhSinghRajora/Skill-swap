'use client';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import Image from 'next/image';
import { useState, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { Check, Circle } from 'lucide-react';
import api, { setTokens, notifyAuthChange, ApiClientError } from '@/lib/api';
import { initializeE2EE } from '@/lib/e2ee/init';

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

  const passwordError = !allChecksMet ? 'Please meet all password requirements' : null;
  if (passwordError) {
    setError(passwordError);
    return;
  }

  setLoading(true);

  try {
    const data = await api.auth.register({
      name: userData.name,
      email: userData.email,
      password: userData.password,
    });

    setTokens(data.access_token, data.refresh_token);
    localStorage.setItem('user', JSON.stringify(data.user));
    notifyAuthChange();

    // Initialize E2EE keys (generate new pair + backup)
    // Fire-and-forget: don't block navigation if E2EE init fails
    initializeE2EE(userData.password).catch(() => {});

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
    <div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card className="overflow-hidden p-0">
        <CardContent className="grid p-0 md:grid-cols-2">
          <form className="p-6 md:p-8" onSubmit={handleSignup}>
            <div className="flex flex-col gap-6">
              <div className="flex flex-col items-center text-center">
                <h1 className="text-2xl font-bold">Create an account</h1>
                <p className="text-muted-foreground">Sign up for SkillSwap</p>
              </div>

              <div className="grid gap-3">
                <Label htmlFor="name">Full Name</Label>
                <Input id="name" required placeholder="John Doe"
                  value={userData.name}
                  onChange={(e) => setUserData({ ...userData, name: e.target.value })}
                />
              </div>

              <div className="grid gap-3">
                <Label htmlFor="email">Email</Label>
                <Input id="email" type="email" required placeholder="abc@example.com"
                  value={userData.email}
                  onChange={(e) => setUserData({ ...userData, email: e.target.value })}
                />
              </div>

              <div className="grid gap-3">
                <Label htmlFor="password">Password</Label>
                <Input id="password" type="password" required
                  value={userData.password}
                  onChange={(e) => setUserData({ ...userData, password: e.target.value })}
                  placeholder="Create a strong password"
                />
                {userData.password.length > 0 && (
                  <ul className="mt-1 space-y-1.5">
                    {passwordChecks.map((check) => (
                      <li key={check.label} className="flex items-center gap-2 text-xs">
                        {check.met ? (
                          <Check className="h-3.5 w-3.5 text-green-500 shrink-0" />
                        ) : (
                          <Circle className="h-3.5 w-3.5 text-muted-foreground/40 shrink-0" />
                        )}
                        <span className={check.met ? 'text-green-600 dark:text-green-400' : 'text-muted-foreground'}>
                          {check.label}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              {error && <p className="text-red-500 text-sm">{error}</p>}

              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? 'Creating Account...' : 'Sign up'}
              </Button>

              <div className="text-center text-sm">
                Already have an account?{' '}
                <a href="/auth/signin" className="underline underline-offset-4">Login</a>
              </div>
            </div>
          </form>
          <div className="bg-muted relative hidden md:block">
            <Image
              src="https://media.gettyimages.com/id/1125868664/video/making-smart-moves-across-the-digital-landscape.jpg?s=640x640&k=20&c=94zKlF9shOT1fiQXShCgJHB2X2_AkYAjTJ4tsj-3uTs="
              alt="Image"
              width={500}
              height={500}
              className="absolute inset-0 h-full w-full object-cover dark:brightness-[0.2] dark:grayscale"
            />
          </div>
        </CardContent>
      </Card>
      <div className="text-muted-foreground text-center text-xs">
        By continuing, you agree to our <a href="#">Terms</a> & <a href="#">Privacy Policy</a>.
      </div>
    </div>
  );
}
