'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar } from '@/components/ui/avatar';
import { MessageSquare, Star, TrendingUp, Calendar, ArrowRight, Plus, Search } from 'lucide-react';
import Link from 'next/link';
import { getUserById, getSwapRequestsByUserId, getNotificationsByUserId } from '@/lib/dummy-data';

export default function DashboardPage() {
	const currentUserId = '1'; // Sarah Johnson
	const user = getUserById(currentUserId);
	const swapRequests = getSwapRequestsByUserId(currentUserId);
	const notifications = getNotificationsByUserId();

	if (!user) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<div className="text-center">
					<h1 className="text-2xl font-bold text-foreground mb-2">User not found</h1>
					<p className="text-muted-foreground">Unable to load dashboard.</p>
				</div>
			</div>
		);
	}

	const pendingSwaps = swapRequests.filter((req) => req.status === 'pending');
	const completedSwaps = swapRequests.filter((req) => req.status === 'accepted');
	const unreadNotifications = notifications.filter((n) => !n.isRead);

	const stats = [
		{
			title: 'Total Swaps',
			value: swapRequests.length,
			icon: TrendingUp,
			description: 'All time exchanges',
			color: 'text-blue-600'
		},
		{
			title: 'Completed Swaps',
			value: completedSwaps.length,
			icon: Star,
			description: 'Successful exchanges',
			color: 'text-green-600'
		},
		{
			title: 'Pending Requests',
			value: pendingSwaps.length,
			icon: Calendar,
			description: 'Awaiting response',
			color: 'text-yellow-600'
		},
		{
			title: 'New Messages',
			value: unreadNotifications.filter((n) => n.type === 'messageReceived').length,
			icon: MessageSquare,
			description: 'Unread conversations',
			color: 'text-purple-600'
		}
	];

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Welcome Header */}
				<div className="mb-8">
					<h1 className="text-3xl font-bold text-foreground mb-2">
						Welcome back, {user.name.split(' ')[0]}!
					</h1>
					<p className="text-muted-foreground">
						Here&apos;s what&apos;s happening with your skill exchanges
					</p>
				</div>

				{/* Stats Grid */}
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
					{stats.map((stat) => {
						const Icon = stat.icon;
						return (
							<Card key={stat.title}>
								<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
									<CardTitle className="text-sm font-medium">
										{stat.title}
									</CardTitle>
									<Icon className={`w-4 h-4 ${stat.color}`} />
								</CardHeader>
								<CardContent>
									<div className="text-2xl font-bold">{stat.value}</div>
									<p className="text-xs text-muted-foreground">
										{stat.description}
									</p>
								</CardContent>
							</Card>
						);
					})}
				</div>

				<div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
					{/* Recent Activity */}
					<div className="lg:col-span-2">
						<Card>
							<CardHeader>
								<CardTitle>Recent Activity</CardTitle>
								<CardDescription>
									Your latest skill exchanges and interactions
								</CardDescription>
							</CardHeader>
							<CardContent>
								<div className="space-y-4">
									{swapRequests.slice(0, 5).map((request) => {
										const otherUser =
											request.requesterId === currentUserId
												? getUserById(request.responderId)
												: getUserById(request.requesterId);

										return (
											<div
												key={request.swapId}
												className="flex items-center space-x-4 p-3 border border-border rounded-lg"
											>
												<Avatar
													src={otherUser?.photoUrl || undefined}
													alt={otherUser?.name || 'User'}
													fallback={
														otherUser?.name
															?.split(' ')
															.map((n) => n[0])
															.join('') || 'U'
													}
													className="w-10 h-10"
												/>
												<div className="flex-1">
													<p className="font-medium">
														{request.requesterId === currentUserId
															? 'You offered'
															: `${otherUser?.name} offered`}{' '}
														{request.offeredSkill.name}
													</p>
													<p className="text-sm text-muted-foreground">
														for {request.wantedSkill.name} •{' '}
														{new Date(
															request.createdAt
														).toLocaleDateString()}
													</p>
												</div>
												<Badge
													variant={
														request.status === 'accepted'
															? 'default'
															: request.status === 'rejected'
																? 'destructive'
																: 'secondary'
													}
												>
													{request.status.charAt(0).toUpperCase() +
														request.status.slice(1)}
												</Badge>
											</div>
										);
									})}
								</div>
								<div className="mt-4">
									<Link href="/profile">
										<Button variant="outline" className="w-full">
											View All Activity
											<ArrowRight className="w-4 h-4 ml-2" />
										</Button>
									</Link>
								</div>
							</CardContent>
						</Card>
					</div>

					{/* Quick Actions & Notifications */}
					<div className="space-y-6">
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
										<MessageSquare className="w-4 h-4 mr-2" />
										View Messages
									</Button>
								</Link>
								<Link href="/profile">
									<Button variant="outline" className="w-full justify-start">
										<Plus className="w-4 h-4 mr-2" />
										Update Skills
									</Button>
								</Link>
							</CardContent>
						</Card>

						{/* Recent Notifications */}
						<Card>
							<CardHeader>
								<CardTitle>Recent Notifications</CardTitle>
							</CardHeader>
							<CardContent>
								<div className="space-y-3">
									{notifications.slice(0, 3).map((notification) => (
										<div
											key={notification.notificationId}
											className="flex items-start space-x-3"
										>
											<div
												className={`w-2 h-2 rounded-full mt-2 ${
													notification.isRead ? 'bg-muted' : 'bg-primary'
												}`}
											/>
											<div className="flex-1">
												<p className="text-sm font-medium">
													{notification.content}
												</p>
												<p className="text-xs text-muted-foreground">
													{new Date(
														notification.createdAt
													).toLocaleDateString()}
												</p>
											</div>
										</div>
									))}
								</div>
								{notifications.length > 3 && (
									<Button variant="outline" className="w-full mt-3">
										View All Notifications
									</Button>
								)}
							</CardContent>
						</Card>

						{/* Skills Summary */}
						<Card>
							<CardHeader>
								<CardTitle>Your Skills</CardTitle>
							</CardHeader>
							<CardContent>
								<div className="space-y-3">
									<div>
										<h4 className="text-sm font-medium mb-2">
											Skills Offered ({user.skillsOffered.length})
										</h4>
										<div className="flex flex-wrap gap-1">
											{user.skillsOffered
												.filter((skill) => skill)
												.map((skill) => (
													<Badge
														key={skill.skillId}
														variant="secondary"
														className="text-xs"
													>
														{skill.name}
													</Badge>
												))}
										</div>
									</div>
									<div>
										<h4 className="text-sm font-medium mb-2">
											Skills Wanted ({user.skillsWanted.length})
										</h4>
										<div className="flex flex-wrap gap-1">
											{user.skillsWanted
												.filter((skill) => skill)
												.map((skill) => (
													<Badge
														key={skill.skillId}
														variant="outline"
														className="text-xs"
													>
														{skill.name}
													</Badge>
												))}
										</div>
									</div>
								</div>
							</CardContent>
						</Card>
					</div>
				</div>

				{/* Popular Skills in Your Area */}
				<Card className="mt-8">
					<CardHeader>
						<CardTitle>Popular Skills in Your Area</CardTitle>
						<CardDescription>
							Skills that are trending among users near you
						</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
							{[
								'JavaScript',
								'Cooking',
								'Photography',
								'Spanish',
								'Guitar',
								'Yoga'
							].map((skill) => (
								<div
									key={skill}
									className="text-center p-3 border border-border rounded-lg hover:border-primary/50 transition-colors cursor-pointer"
								>
									<h3 className="font-medium text-sm">{skill}</h3>
									<p className="text-xs text-muted-foreground mt-1">
										{Math.floor(Math.random() * 20) + 5} people
									</p>
								</div>
							))}
						</div>
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
