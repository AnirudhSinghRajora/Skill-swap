'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { users, type UserProfileResponse } from '@/lib/api';
import { Navigation } from '@/components/Navigation';

export default function PublicSlugProfilePage() {
  const params = useParams<{ slug: string }>();
  const slug = params?.slug ?? '';
  const [profile, setProfile] = useState<UserProfileResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    users
      .getBySlug(slug)
      .then((p) => {
        if (!cancelled) setProfile(p);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Not found');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [slug]);

  return (
    <main className="min-h-screen bg-background">
      <Navigation />
      <div className="container mx-auto max-w-3xl px-4 py-12">
        {loading && <p className="text-muted-foreground">Loading…</p>}
        {error && !loading && (
          <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-center">
            <h1 className="text-xl font-semibold">Profile unavailable</h1>
            <p className="mt-2 text-sm text-muted-foreground">{error}</p>
          </div>
        )}
        {profile && (
          <article className="space-y-6">
            <header className="space-y-1">
              <h1 className="text-3xl font-bold">{profile.name}</h1>
              {profile.location && (
                <p className="text-muted-foreground">{profile.location}</p>
              )}
              {profile.slug && (
                <p className="text-xs text-muted-foreground">@{profile.slug}</p>
              )}
            </header>
            <section>
              <h2 className="mb-2 text-lg font-semibold">Offers</h2>
              <div className="flex flex-wrap gap-2">
                {profile.skills_offered.length === 0 && (
                  <span className="text-sm text-muted-foreground">No skills offered</span>
                )}
                {profile.skills_offered.map((s) => (
                  <span
                    key={s.skill_id}
                    className="rounded-full bg-primary/10 px-3 py-1 text-xs font-medium text-primary"
                  >
                    {s.name}
                  </span>
                ))}
              </div>
            </section>
            <section>
              <h2 className="mb-2 text-lg font-semibold">Wants to learn</h2>
              <div className="flex flex-wrap gap-2">
                {profile.skills_wanted.length === 0 && (
                  <span className="text-sm text-muted-foreground">No skills wanted</span>
                )}
                {profile.skills_wanted.map((s) => (
                  <span
                    key={s.skill_id}
                    className="rounded-full bg-secondary px-3 py-1 text-xs font-medium"
                  >
                    {s.name}
                  </span>
                ))}
              </div>
            </section>
          </article>
        )}
      </div>
    </main>
  );
}
