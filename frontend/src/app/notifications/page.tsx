'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Bell, Check, CheckCheck, Trash2, Loader2 } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '@/lib/api';
import { toast } from 'sonner';
import Link from 'next/link';
import type { NotificationType } from '@/types/notification';

function getNotificationHref(type: NotificationType): string {
	switch (type) {
		case 'swap_request':
		case 'swap_accepted':
		case 'swap_rejected':
		case 'swap_completed':
			return '/swaps';
		case 'new_rating':
			return '/profile?tab=reviews';
		case 'skill_matched':
			return '/browse';
		default:
			return '/notifications';
	}
}

export default function NotificationsPage() {
	const { user, isLoading: authLoading } = useAuth(true);
	const queryClient = useQueryClient();

	const { data: notifData, isLoading: notifsLoading } = useQuery({
		queryKey: ['notifications'],
		queryFn: () => api.notifications.list({ limit: 50 }),
		enabled: !!user,
	});

	const { data: stats } = useQuery({
		queryKey: ['notification-stats'],
		queryFn: () => api.notifications.stats(),
		enabled: !!user,
	});

	const markReadMutation = useMutation({
		mutationFn: (ids: string[]) => api.notifications.markRead(ids),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['notifications'] });
			queryClient.invalidateQueries({ queryKey: ['notification-stats'] });
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to mark as read'),
	});

	const markAllReadMutation = useMutation({
		mutationFn: () => api.notifications.markAllRead(),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['notifications'] });
			queryClient.invalidateQueries({ queryKey: ['notification-stats'] });
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to mark all as read'),
	});

	const deleteMutation = useMutation({
		mutationFn: (id: string) => api.notifications.delete(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['notifications'] });
			queryClient.invalidateQueries({ queryKey: ['notification-stats'] });
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to delete notification'),
	});

	if (authLoading || notifsLoading) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
			</div>
		);
	}

	const notifications = notifData?.notifications ?? [];
	const unreadCount = stats?.unread_count ?? 0;

	const getTypeColor = (type: string) => {
		switch (type) {
			case 'swap_request':
				return 'bg-blue-500';
			case 'swap_accepted':
				return 'bg-green-500';
			case 'swap_rejected':
				return 'bg-red-500';
			case 'new_rating':
				return 'bg-yellow-500';
			case 'skill_matched':
				return 'bg-purple-500';
			default:
				return 'bg-muted-foreground';
		}
	};

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Header */}
				<div className="flex items-center justify-between mb-10 animate-fade-in-up">
					<div>
						<h1 className="text-display-md text-foreground mb-2">Notifications</h1>
						<p className="text-muted-foreground text-lg">
							{unreadCount > 0
								? `You have ${unreadCount} unread notification${unreadCount === 1 ? '' : 's'}`
								: 'All caught up!'}
						</p>
					</div>
					{unreadCount > 0 && (
						<Button
							variant="outline"
							size="sm"
							onClick={() => markAllReadMutation.mutate()}
							disabled={markAllReadMutation.isPending}
						>
							{markAllReadMutation.isPending ? (
								<Loader2 className="w-4 h-4 mr-2 animate-spin" />
							) : (
								<CheckCheck className="w-4 h-4 mr-2" />
							)}
							Mark All Read
						</Button>
					)}
				</div>

				{/* Notifications List */}
				{notifications.length > 0 ? (
					<div className="space-y-3">
						{notifications.map((notification) => (
							<Card
								key={notification.notification_id}
								className={notification.is_read ? 'opacity-70' : ''}
							>
								<CardContent className="flex items-start gap-4 p-4">
									<Link
										href={getNotificationHref(notification.type)}
										className="flex items-start gap-4 flex-1 min-w-0 hover:opacity-80 transition-opacity"
										onClick={() => {
											if (!notification.is_read) {
												markReadMutation.mutate([notification.notification_id]);
											}
										}}
									>
										<div
											className={`w-2.5 h-2.5 rounded-full mt-2 shrink-0 ${
												notification.is_read
													? 'bg-muted'
													: getTypeColor(notification.type)
											}`}
										/>
										<div className="flex-1 min-w-0">
											<p className="font-medium text-sm">
												{notification.title}
											</p>
											<p className="text-sm text-muted-foreground mt-1">
												{notification.message}
											</p>
											<p className="text-xs text-muted-foreground mt-2">
												{new Date(
													notification.created_at
												).toLocaleString()}
											</p>
										</div>
									</Link>
									<div className="flex items-center gap-1 shrink-0">
										{!notification.is_read && (
											<Button
												size="sm"
												variant="ghost"
												onClick={() =>
													markReadMutation.mutate([
														notification.notification_id,
													])
												}
												title="Mark as read"
											>
												<Check className="w-4 h-4" />
											</Button>
										)}
										<Button
											size="sm"
											variant="ghost"
											onClick={() =>
												deleteMutation.mutate(
													notification.notification_id
												)
											}
											title="Delete"
										>
											<Trash2 className="w-4 h-4 text-destructive" />
										</Button>
									</div>
								</CardContent>
							</Card>
						))}
					</div>
				) : (
					<Card>
						<CardContent className="flex flex-col items-center py-12">
							<Bell className="w-12 h-12 text-muted-foreground mb-4" />
							<h3 className="text-lg font-medium text-foreground mb-1">
								No notifications
							</h3>
							<p className="text-muted-foreground">
								You&apos;ll see notifications here when something happens
							</p>
						</CardContent>
					</Card>
				)}
			</div>
		</div>
	);
}
