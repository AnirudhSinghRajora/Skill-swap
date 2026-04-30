'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '@/lib/api';
import { Card, CardHeader, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import { ShieldCheck, Loader2 } from 'lucide-react';

const STATUSES = ['open', 'reviewing', 'resolved', 'dismissed'] as const;

export default function AdminAbuseReportsPage() {
  const [status, setStatus] = useState<string>('open');
  const qc = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin', 'abuse-reports', status],
    queryFn: () => api.moderation.adminListReports({ status, limit: 100 }),
    retry: false,
  });

  const resolve = useMutation({
    mutationFn: ({ id, status, note }: { id: string; status: 'resolved' | 'dismissed' | 'reviewing'; note?: string }) =>
      api.moderation.adminResolveReport(id, { status, note }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'abuse-reports'] });
      toast.success('Report updated');
    },
    onError: (err: Error) => toast.error(err.message),
  });

  if (error) {
    return (
      <div className="container max-w-4xl mx-auto py-12">
        <Card>
          <CardContent className="py-10 text-center">
            <p className="text-muted-foreground">
              {(error as Error).message.includes('403') || (error as Error).message.toLowerCase().includes('forbidden')
                ? 'Access denied. Admins only.'
                : (error as Error).message}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container max-w-5xl mx-auto py-8 px-4">
      <div className="flex items-center gap-3 mb-6">
        <ShieldCheck className="w-6 h-6" />
        <h1 className="text-2xl font-bold">Abuse Reports</h1>
      </div>

      <div className="flex gap-2 mb-6">
        {STATUSES.map((s) => (
          <Button
            key={s}
            variant={status === s ? 'default' : 'outline'}
            size="sm"
            onClick={() => setStatus(s)}
          >
            {s}
          </Button>
        ))}
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="w-6 h-6 animate-spin" />
        </div>
      ) : data && data.reports.length > 0 ? (
        <div className="space-y-3">
          {data.reports.map((r) => (
            <Card key={r.report_id}>
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between gap-3">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2 text-sm">
                      <Badge variant="outline">{r.target_kind}</Badge>
                      <Badge>{r.status}</Badge>
                      <span className="text-muted-foreground">
                        {new Date(r.created_at).toLocaleString()}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground font-mono">
                      Reporter: {r.reporter_id.slice(0, 8)}… • Target: {r.target_user_id.slice(0, 8)}…
                      {r.target_id && <> • Item: {r.target_id.slice(0, 8)}…</>}
                    </p>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm whitespace-pre-wrap">{r.reason}</p>
                {r.resolution_note && (
                  <p className="text-xs text-muted-foreground border-l-2 border-muted pl-3">
                    Note: {r.resolution_note}
                  </p>
                )}
                {(r.status === 'open' || r.status === 'reviewing') && (
                  <div className="flex gap-2">
                    {r.status === 'open' && (
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={resolve.isPending}
                        onClick={() => resolve.mutate({ id: r.report_id, status: 'reviewing' })}
                      >
                        Mark reviewing
                      </Button>
                    )}
                    <Button
                      size="sm"
                      disabled={resolve.isPending}
                      onClick={() => {
                        const note = window.prompt('Resolution note (optional):') || '';
                        resolve.mutate({ id: r.report_id, status: 'resolved', note });
                      }}
                    >
                      Resolve
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={resolve.isPending}
                      onClick={() => {
                        const note = window.prompt('Dismissal note (optional):') || '';
                        resolve.mutate({ id: r.report_id, status: 'dismissed', note });
                      }}
                    >
                      Dismiss
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <Card>
          <CardContent className="py-10 text-center text-muted-foreground">
            No reports with status &ldquo;{status}&rdquo;.
          </CardContent>
        </Card>
      )}
    </div>
  );
}
