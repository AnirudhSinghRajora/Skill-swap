'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar } from '@/components/ui/avatar';
import { Star, TrendingUp, Calendar, ArrowRight, Search, Loader2, MessageCircle } from 'lucide-react';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { useQuery } from '@tanstack/react-query';
import api, { getPhotoUrl } from '@/lib/api';
import { Skeleton } from '@/components/ui/skeleton';
import type { NotificationType } from '@/types/notification';

function getNotificationHref(notification: { type: NotificationType; related_id: string | null }): string {
	switch (notification.type) {
		case 'swap_request':
		case 'swap_accepted':
		case 'swap_rejected':
		case 'swap_completed':
			return '/swaps';
		case 'new_rating':
			return '/profile?tab=reviews';
		case 'new_message':
			return notification.related_id
				? `/messages?conversation=${notification.related_id}`
				: '/messages';
		case 'skill_matched':
			return '/browse';
		default:
			return '/notifications';
	}
}

export default function DashboardPage() {
	const { user, isLoading: authLoading } = useAuth(true);

	const { data: profile, isLoading: profileLoading } = useQuery({
		queryKey: ['profile'],
		queryFn: () => api.users.getProfile(),
		enabled: !!user,
	});

	const { data: swapData, isLoading: swapsLoading } = useQuery({
		queryKey: ['swaps'],
		queryFn: () => api.swaps.list(),
		enabled: !!user,
	});

	const { data: notifData } = useQuery({
		queryKey: ['notifications-recent'],
		queryFn: () => api.notifications.list({ limit: 5 }),
		enabled: !!user,
	});

	const { data: ratingStats } = useQuery({
		queryKey: ['rating-stats', user?.user_id],
		queryFn: () => api.ratings.getStatsForUser(user!.user_id),
		enabled: !!user?.user_id,
	});

	const { data: conversationsData } = useQuery({
		queryKey: ['conversations'],
		queryFn: () => api.conversations.list(),
		enabled: !!user,
	});

	const isDataLoading = authLoading || profileLoading || swapsLoading;

	if (authLoading) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
			</div>
		);
	}

	const allSwaps = [...(swapData?.sent ?? []), ...(swapData?.received ?? [])];
	const pendingSwaps = allSwaps.filter((s) => s.status === 'pending');
	const activeSwaps = allSwaps.filter((s) => s.status === 'accepted');
	const completedSwaps = allSwaps.filter((s) => s.status === 'completed');
	const recentSwaps = allSwaps
		.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
		.slice(0, 5);
	const notifications = notifData?.notifications ?? [];
	const unreadCount = notifications.filter((n) => !n.is_read).length;
	const conversations = conversationsData?.conversations ?? [];
	const recentConversations = [...conversations]
		.sort((a, b) => {
			const aTime = a.last_message?.created_at ?? a.created_at;
			const bTime = b.last_message?.created_at ?? b.created_at;
			return new Date(bTime).getTime() - new Date(aTime).getTime();
		})
		.slice(0, 3);

	const stats = [
		{
			title: 'Total Swaps',
			value: allSwaps.length,
			icon: TrendingUp,
			description: 'All time exchanges',
			color: 'text-blue-600',
		},
		{
			title: 'In Progress',
			value: activeSwaps.length,
			icon: MessageCircle,
			description: `${completedSwaps.length} completed`,
			color: 'text-green-600',
		},
		{
			title: 'Pending',
			value: pendingSwaps.length,
			icon: Calendar,
			description: 'Awaiting response',
			color: 'text-yellow-600',
		},
		{
			title: 'Avg Rating',
			value: ratingStats?.average_rating ? ratingStats.average_rating.toFixed(1) : '—',
			icon: Star,
			description: `${ratingStats?.total_ratings ?? 0} reviews`,
			color: 'text-purple-600',
		},
	];

	if (isDataLoading) {
		return (
			<div className="min-h-screen bg-background">
				<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
					<div className="mb-8">
						<Skeleton className="h-9 w-72 mb-2" />
						<Skeleton className="h-5 w-96" />
					</div>
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
						{[1, 2, 3, 4].map((i) => (
							<Card key={i}>
								<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
									<Skeleton className="h-4 w-20" />
									<Skeleton className="h-4 w-4 rounded" />
								</CardHeader>
								<CardContent>
									<Skeleton className="h-8 w-12 mb-1" />
									<Skeleton className="h-3 w-24" />
								</CardContent>
							</Card>
						))}
					</div>
					<div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
						<div className="lg:col-span-2">
							<Card>
								<CardHeader>
									<Skeleton className="h-6 w-36 mb-1" />
									<Skeleton className="h-4 w-64" />
								</CardHeader>
								<CardContent className="space-y-4">
									{[1, 2, 3].map((i) => (
										<div key={i} className="flex items-center gap-3">
											<Skeleton className="h-10 w-10 rounded-full" />
											<div className="flex-1">
												<Skeleton className="h-4 w-48 mb-1" />
												<Skeleton className="h-3 w-32" />
											</div>
										</div>
									))}
								</CardContent>
							</Card>
						</div>
						<Card>
							<CardHeader>
								<Skeleton className="h-6 w-28 mb-1" />
								<Skeleton className="h-4 w-40" />
							</CardHeader>
							<CardContent className="space-y-3">
								{[1, 2, 3].map((i) => (
									<Skeleton key={i} className="h-12 w-full rounded" />
								))}
							</CardContent>
						</Card>
					</div>
				</div>
			</div>
		);
	}

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Welcome Header */}
				<div className="mb-10 animate-fade-in-up">
					<h1 className="text-display-md text-foreground mb-2">
						Welcome back, {profile?.name?.split(' ')[0] ?? 'there'}
					</h1>
					<p className="text-muted-foreground text-lg">
						Here&apos;s what&apos;s happening with your skill exchanges
					</p>
				</div>

				{/* Stats Grid */}
				<div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-12 stagger-children">
					{stats.map((stat, index) => {
						const Icon = stat.icon;
						const isFeatured = index === 0;
						const linkMap: Record<string, string> = {
							'Total Swaps': '/swaps',
							'In Progress': '/swaps',
							'Pending': '/swaps',
							'Avg Rating': `/profile?tab=reviews`,
						};
						return (
							<Link key={stat.title} href={linkMap[stat.title] ?? '/dashboard'} className={isFeatured ? 'col-span-2 lg:col-span-1' : ''}>
								<Card className={`transition-colors cursor-pointer h-full ${
									isFeatured
										? 'bg-primary/5 border-primary/20 hover:bg-primary/10'
										: 'hover:bg-muted/50'
								}`}>
									<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
										<CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
										<Icon className={`w-4 h-4 ${isFeatured ? 'text-primary' : stat.color}`} />
									</CardHeader>
									<CardContent>
										<div className={`font-bold ${isFeatured ? 'text-3xl' : 'text-2xl'}`}>{stat.value}</div>
										<p className="text-xs text-muted-foreground">{stat.description}</p>
									</CardContent>
								</Card>
							</Link>
						);
					})}
				</div>

				<div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
					{/* Recent Activity */}
					<div className="lg:col-span-2">
						<Card>
							<CardHeader>
								<CardTitle>Recent Activity</CardTitle>
								<CardDescription>Your latest skill exchanges and interactions</CardDescription>
							</CardHeader>
							<CardContent>
								{recentSwaps.length > 0 ? (
									<div className="space-y-4">
										{recentSwaps.map((request) => {
											const isRequester = request.requester_id === user?.user_id;
											const other = isRequester ? request.responder : request.requester;

											return (
												<Link
													key={request.swap_id}
													href="/swaps"
													className="flex items-center space-x-4 p-3 border border-border rounded-lg hover:bg-muted/50 transition-colors"
												>
													<Avatar
													src={other.has_photo ? getPhotoUrl(other.user_id) : undefined}
														alt={other.name}
														fallback={
															other.name
																.split(' ')
																.map((n) => n[0])
																.join('') || 'U'
														}
														className="w-10 h-10"
													/>
													<div className="flex-1">
														<p className="font-medium">
															{isRequester ? 'You offered' : `${other.name} offered`}{' '}
															{request.offered_skill.name}
														</p>
														<p className="text-sm text-muted-foreground">
															for {request.wanted_skill.name} •{' '}
															{new Date(request.created_at).toLocaleDateString()}
														</p>
													</div>
													<Badge
														variant={
															request.status === 'accepted'
																? 'default'
																: request.status === 'completed'
																	? 'outline'
																	: request.status === 'rejected' || request.status === 'cancelled'
																		? 'destructive'
																		: 'secondary'
														}
													>
														{request.status === 'accepted' ? 'In Progress' : request.status === 'cancelled' ? 'Cancelled' : request.status.charAt(0).toUpperCase() + request.status.slice(1)}
													</Badge>
												</Link>
											);
										})}
									</div>
								) : (
								<div className="text-center py-12">
									<div className="w-14 h-14 mx-auto mb-4 rounded-2xl bg-muted flex items-center justify-center">
										<TrendingUp className="w-7 h-7 text-muted-foreground" />
									</div>
									<h3 className="font-medium text-foreground mb-1">No swaps yet</h3>
									<p className="text-sm text-muted-foreground">Browse users to find your first skill exchange</p>
									</div>
								)}
								<div className="mt-4">
									<Link href="/swaps">
										<Button variant="outline" className="w-full">
											View All Swaps
											<ArrowRight className="w-4 h-4 ml-2" />
										</Button>
									</Link>
								</div>
							</CardContent>
						</Card>
					</div>

					{/* Right sidebar */}
					<div className="space-y-6">
						{/* Recent Messages */}
						{recentConversations.length > 0 && (
							<Card>
								<CardHeader className="flex flex-row items-center justify-between space-y-0">
									<CardTitle>Recent Messages</CardTitle>
									<Link href="/messages">
										<Button variant="ghost" size="sm" className="text-xs">
											View all
										</Button>
									</Link>
								</CardHeader>
								<CardContent>
									<div className="space-y-3">
										{recentConversations.map((convo) => (
											<Link
												key={convo.conversation_id}
												href={`/messages?conversation=${convo.conversation_id}`}
												className="flex items-center gap-3 p-2 -mx-2 rounded-md hover:bg-muted/50 transition-colors"
											>
												<Avatar
													src={convo.other_user.has_photo ? getPhotoUrl(convo.other_user.user_id) : undefined}
													alt={convo.other_user.name}
													fallback={convo.other_user.name.charAt(0).toUpperCase()}
													className="w-8 h-8 shrink-0"
												/>
												<div className="min-w-0 flex-1">
													<p className="text-sm font-medium truncate">{convo.other_user.name}</p>
													{convo.last_message ? (
														<p className="text-xs text-muted-foreground truncate">
															{convo.last_message.content.replace(/<[^>]*>/g, '').slice(0, 50)}
														</p>
													) : (
														<p className="text-xs text-muted-foreground italic">No messages yet</p>
													)}
												</div>
												{convo.unread_count > 0 && (
													<Badge variant="destructive" className="text-[10px] h-5 min-w-5 flex items-center justify-center">
														{convo.unread_count}
													</Badge>
												)}
											</Link>
										))}
									</div>
								</CardContent>
							</Card>
						)}

						{/* Quick Actions */}
						<Card>
							<CardHeader>
								<CardTitle>Quick Actions</CardTitle>
							</CardHeader>
							<CardContent className="space-y-3">
								<Link href="/browse">
									<Button className="w-full justify-start">
										<Search className="w-4 h-4 mr-2" />
										Find People
									</Button>
								</Link>
								<Link href="/messages">
									<Button variant="outline" className="w-full justify-start">
										<MessageCircle className="w-4 h-4 mr-2" />
										Messages
									</Button>
								</Link>
								<Link href="/swaps">
									<Button variant="outline" className="w-full justify-start">
										<ArrowRight className="w-4 h-4 mr-2" />
										Manage Swaps
									</Button>
								</Link>
								<Link href="/profile">
									<Button variant="outline" className="w-full justify-start">
										<Star className="w-4 h-4 mr-2" />
										Update Skills
									</Button>
								</Link>
							</CardContent>
						</Card>

						{/* Recent Notifications */}
						<Card>
							<CardHeader className="flex flex-row items-center justify-between space-y-0">
								<CardTitle>Notifications</CardTitle>
								{unreadCount > 0 && (
									<Badge variant="destructive" className="text-xs">
										{unreadCount} new
									</Badge>
								)}
							</CardHeader>
							<CardContent>
								{notifications.length > 0 ? (
									<div className="space-y-3">
										{notifications.slice(0, 4).map((notification) => (
											<Link
												key={notification.notification_id}
											href={getNotificationHref(notification)}
												className="flex items-start space-x-3 hover:bg-muted/50 rounded-md p-1 -m-1 transition-colors"
											>
												<div
													className={`w-2 h-2 rounded-full mt-2 ${
														notification.is_read ? 'bg-muted' : 'bg-primary'
													}`}
												/>
												<div className="flex-1">
													<p className="text-sm font-medium">{notification.title}</p>
													<p className="text-xs text-muted-foreground">
														{notification.message}
													</p>
													<p className="text-xs text-muted-foreground">
														{new Date(notification.created_at).toLocaleDateString()}
													</p>
												</div>
											</Link>
										))}
									</div>
								) : (
									<p className="text-sm text-muted-foreground text-center py-4">
										No notifications yet
									</p>
								)}
							</CardContent>
						</Card>

						{/* Your Skills */}
						{profile && (
							<Card>
								<CardHeader>
									<CardTitle>Your Skills</CardTitle>
								</CardHeader>
								<CardContent>
									<div className="space-y-3">
										<div>
											<h4 className="text-sm font-medium mb-2">
												Offering ({profile.skills_offered.length})
											</h4>
											<div className="flex flex-wrap gap-1">
												{profile.skills_offered.length > 0 ? (
													profile.skills_offered.map((skill) => (
														<Badge key={skill.skill_id} variant="secondary" className="text-xs">
															{skill.name}
														</Badge>
													))
												) : (
													<p className="text-xs text-muted-foreground">None yet</p>
												)}
											</div>
										</div>
										<div>
											<h4 className="text-sm font-medium mb-2">
												Looking for ({profile.skills_wanted.length})
											</h4>
											<div className="flex flex-wrap gap-1">
												{profile.skills_wanted.length > 0 ? (
													profile.skills_wanted.map((skill) => (
														<Badge key={skill.skill_id} variant="outline" className="text-xs">
															{skill.name}
														</Badge>
													))
												) : (
													<p className="text-xs text-muted-foreground">None yet</p>
												)}
											</div>
										</div>
									</div>
								</CardContent>
							</Card>
						)}
					</div>
				</div>
			</div>
		</div>
	);
}
