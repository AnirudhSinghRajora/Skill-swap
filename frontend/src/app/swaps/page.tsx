'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Avatar } from '@/components/ui/avatar';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { CheckCircle, XCircle, Clock, MessageSquare, ArrowRight, User } from 'lucide-react';
import { getUserById, getSwapRequestsByUserId } from '@/lib/dummy-data';
import { SwapRequest } from '@/types/swapRequest';

export default function SwapsPage() {
	const [activeTab, setActiveTab] = useState('incoming');
	const currentUserId = '1'; // Sarah Johnson
	const user = getUserById(currentUserId);
	const allSwapRequests = getSwapRequestsByUserId(currentUserId);

	if (!user) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<div className="text-center">
					<h1 className="text-2xl font-bold text-foreground mb-2">User not found</h1>
					<p className="text-muted-foreground">Unable to load swap requests.</p>
				</div>
			</div>
		);
	}

	const incomingRequests = allSwapRequests.filter((req) => req.responderId === currentUserId);
	const outgoingRequests = allSwapRequests.filter((req) => req.requesterId === currentUserId);

	const getStatusIcon = (status: string) => {
		switch (status) {
			case 'accepted':
				return <CheckCircle className="w-4 h-4 text-green-500" />;
			case 'rejected':
				return <XCircle className="w-4 h-4 text-red-500" />;
			case 'pending':
				return <Clock className="w-4 h-4 text-yellow-500" />;
			default:
				return null;
		}
	};

	const handleAcceptRequest = (swapId: string) => {
		console.log('Accepting swap request:', swapId);
		// In a real app, this would update the swap request status
	};

	const handleRejectRequest = (swapId: string) => {
		console.log('Rejecting swap request:', swapId);
		// In a real app, this would update the swap request status
	};

	const handleCancelRequest = (swapId: string) => {
		console.log('Canceling swap request:', swapId);
		// In a real app, this would update the swap request status
	};

	const SwapRequestCard = ({ request, isIncoming }: { request: SwapRequest; isIncoming: boolean }) => {
		const otherUser = isIncoming
			? getUserById(request.requesterId)
			: getUserById(request.responderId);

		return (
			<Card className="hover:shadow-md transition-shadow duration-200">
				<CardHeader>
					<div className="flex items-start justify-between">
						<div className="flex items-center space-x-3">
							<Avatar
								src={otherUser?.photoUrl || undefined}
								alt={otherUser?.name || 'User'}
								fallback={
									otherUser?.name
										?.split(' ')
										.map((n) => n[0])
										.join('') || 'U'
								}
								className="w-12 h-12"
							/>
							<div>
								<h3 className="font-semibold text-lg">{otherUser?.name}</h3>
								<p className="text-sm text-muted-foreground">
									{new Date(request.createdAt).toLocaleDateString()}
								</p>
							</div>
						</div>
						<div className="flex items-center space-x-2">
							{getStatusIcon(request.status)}
							<Badge
								variant={
									request.status === 'accepted'
										? 'default'
										: request.status === 'rejected'
											? 'destructive'
											: 'secondary'
								}
							>
								{request.status.charAt(0).toUpperCase() + request.status.slice(1)}
							</Badge>
						</div>
					</div>
				</CardHeader>

				<CardContent className="space-y-4">
					<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div className="p-3 bg-muted rounded-lg">
							<p className="text-sm font-medium text-muted-foreground mb-1">
								{isIncoming ? 'They Offer' : 'You Offer'}
							</p>
							<p className="font-medium">{request.offeredSkill.name}</p>
							{request.offeredSkill.description && (
								<p className="text-xs text-muted-foreground mt-1">
									{request.offeredSkill.description}
								</p>
							)}
						</div>
						<div className="p-3 bg-muted rounded-lg">
							<p className="text-sm font-medium text-muted-foreground mb-1">
								{isIncoming ? 'They Want' : 'You Want'}
							</p>
							<p className="font-medium">{request.wantedSkill.name}</p>
							{request.wantedSkill.description && (
								<p className="text-xs text-muted-foreground mt-1">
									{request.wantedSkill.description}
								</p>
							)}
						</div>
					</div>

					{/* Actions */}
					<div className="flex space-x-2 pt-2">
						{request.status === 'pending' && isIncoming && (
							<>
								<Button
									onClick={() => handleAcceptRequest(request.swapId)}
									className="flex-1"
								>
									<CheckCircle className="w-4 h-4 mr-2" />
									Accept
								</Button>
								<Button
									variant="outline"
									onClick={() => handleRejectRequest(request.swapId)}
									className="flex-1"
								>
									<XCircle className="w-4 h-4 mr-2" />
									Decline
								</Button>
							</>
						)}
						{request.status === 'pending' && !isIncoming && (
							<Button
								variant="outline"
								onClick={() => handleCancelRequest(request.swapId)}
								className="w-full"
							>
								Cancel Request
							</Button>
						)}
						{request.status === 'accepted' && (
							<Button className="w-full">
								<MessageSquare className="w-4 h-4 mr-2" />
								Start Exchange
							</Button>
						)}
						{request.status === 'rejected' && (
							<Button variant="outline" className="w-full">
								<MessageSquare className="w-4 h-4 mr-2" />
								Message
							</Button>
						)}
					</div>
				</CardContent>
			</Card>
		);
	};

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Header */}
				<div className="mb-8">
					<h1 className="text-3xl font-bold text-foreground mb-2">Swap Requests</h1>
					<p className="text-muted-foreground">Manage your skill exchange requests</p>
				</div>

				{/* Stats */}
				<div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
					<Card>
						<CardContent className="p-4">
							<div className="flex items-center space-x-2">
								<Clock className="w-5 h-5 text-yellow-500" />
								<div>
									<p className="text-2xl font-bold">
										{
											incomingRequests.filter(
												(req) => req.status === 'pending'
											).length
										}
									</p>
									<p className="text-sm text-muted-foreground">Pending</p>
								</div>
							</div>
						</CardContent>
					</Card>
					<Card>
						<CardContent className="p-4">
							<div className="flex items-center space-x-2">
								<CheckCircle className="w-5 h-5 text-green-500" />
								<div>
									<p className="text-2xl font-bold">
										{
											allSwapRequests.filter(
												(req) => req.status === 'accepted'
											).length
										}
									</p>
									<p className="text-sm text-muted-foreground">Accepted</p>
								</div>
							</div>
						</CardContent>
					</Card>
					<Card>
						<CardContent className="p-4">
							<div className="flex items-center space-x-2">
								<XCircle className="w-5 h-5 text-red-500" />
								<div>
									<p className="text-2xl font-bold">
										{
											allSwapRequests.filter(
												(req) => req.status === 'rejected'
											).length
										}
									</p>
									<p className="text-sm text-muted-foreground">Rejected</p>
								</div>
							</div>
						</CardContent>
					</Card>
					<Card>
						<CardContent className="p-4">
							<div className="flex items-center space-x-2">
								<User className="w-5 h-5 text-blue-500" />
								<div>
									<p className="text-2xl font-bold">{allSwapRequests.length}</p>
									<p className="text-sm text-muted-foreground">Total</p>
								</div>
							</div>
						</CardContent>
					</Card>
				</div>

				{/* Tabs */}
				<Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
					<TabsList className="grid w-full grid-cols-2">
						<TabsTrigger value="incoming">
							Incoming Requests ({incomingRequests.length})
						</TabsTrigger>
						<TabsTrigger value="outgoing">
							Outgoing Requests ({outgoingRequests.length})
						</TabsTrigger>
					</TabsList>

					<TabsContent value="incoming" className="space-y-4">
						{incomingRequests.length > 0 ? (
							incomingRequests.map((request) => (
								<SwapRequestCard
									key={request.swapId}
									request={request}
									isIncoming={true}
								/>
							))
						) : (
							<Card>
								<CardContent className="p-8 text-center">
									<div className="w-16 h-16 mx-auto mb-4 bg-muted rounded-full flex items-center justify-center">
										<User className="w-8 h-8 text-muted-foreground" />
									</div>
									<h3 className="text-lg font-medium text-foreground mb-2">
										No incoming requests
									</h3>
									<p className="text-muted-foreground mb-4">
										You don&apos;t have any pending swap requests at the moment.
									</p>
									<Button>
										<ArrowRight className="w-4 h-4 mr-2" />
										Browse People
									</Button>
								</CardContent>
							</Card>
						)}
					</TabsContent>

					<TabsContent value="outgoing" className="space-y-4">
						{outgoingRequests.length > 0 ? (
							outgoingRequests.map((request) => (
								<SwapRequestCard
									key={request.swapId}
									request={request}
									isIncoming={false}
								/>
							))
						) : (
							<Card>
								<CardContent className="p-8 text-center">
									<div className="w-16 h-16 mx-auto mb-4 bg-muted rounded-full flex items-center justify-center">
										<ArrowRight className="w-8 h-8 text-muted-foreground" />
									</div>
									<h3 className="text-lg font-medium text-foreground mb-2">
										No outgoing requests
									</h3>
									<p className="text-muted-foreground mb-4">
										You haven&apos;t sent any swap requests yet.
									</p>
									<Button>
										<ArrowRight className="w-4 h-4 mr-2" />
										Find People to Swap With
									</Button>
								</CardContent>
							</Card>
						)}
					</TabsContent>
				</Tabs>
			</div>
		</div>
	);
}
