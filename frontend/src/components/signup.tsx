'use client';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import Image from 'next/image';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

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

  const API = process.env.NEXT_PUBLIC_API_BASE_URL;

const handleSignup = async (event: React.FormEvent<HTMLFormElement>) => {
  event.preventDefault();
  setError('');
  setLoading(true);

  try {
    const res = await fetch(`${API}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData)
    });

    const data = await res.json();

    if (!res.ok) throw new Error(data?.message || 'Signup failed');

    localStorage.setItem('access_token', data.access_token);
    localStorage.setItem('refresh_token', data.refresh_token);
    localStorage.setItem('user', JSON.stringify(data.user));

    router.push('/');
  } catch (err: unknown) {
    setError(err instanceof Error ? err.message : 'An error occurred');
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
