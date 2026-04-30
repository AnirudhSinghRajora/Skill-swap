'use client';

import { useState, useEffect, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Avatar } from '@/components/ui/avatar';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
	CheckCircle, XCircle, Clock, ArrowRight, User, Loader2, Star,
	X, ArrowRightLeft, MessageCircle, CircleCheckBig, Undo2, Ban,
} from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api, { getPhotoUrl, SwapRequestResponse, UserProfileResponse, SkillResponse, RatingResponse } from '@/lib/api';
import Link from 'next/link';
import { toast } from 'sonner';

function getStatusIcon(status: string) {
	switch (status) {
		case 'accepted':
			return <MessageCircle className="w-4 h-4 text-blue-500" />;
		case 'completed':
			return <CircleCheckBig className="w-4 h-4 text-green-500" />;
		case 'rejected':
			return <XCircle className="w-4 h-4 text-red-500" />;
		case 'cancelled':
			return <Ban className="w-4 h-4 text-gray-400" />;
		case 'pending':
			return <Clock className="w-4 h-4 text-yellow-500" />;
		default:
			return null;
	}
}

function getStatusBadge(status: string) {
	const map: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
		pending:   { label: 'Pending',     variant: 'secondary' },
		accepted:  { label: 'In Progress', variant: 'default' },
		completed: { label: 'Completed',   variant: 'outline' },
		rejected:  { label: 'Rejected',    variant: 'destructive' },
		cancelled: { label: 'Cancelled',   variant: 'secondary' },
	};
	const s = map[status] ?? { label: status, variant: 'secondary' as const };
	return <Badge variant={s.variant}>{s.label}</Badge>;
}

function SwapCard({
	request,
	isIncoming,
	currentUserId,
	onStatusChange,
	statusPending,
}: {
	request: SwapRequestResponse;
	isIncoming: boolean;
	currentUserId: string;
	onStatusChange: (id: string, status: 'accepted' | 'rejected' | 'cancelled') => void;
	statusPending: boolean;
}) {
	const queryClient = useQueryClient();
	const other = isIncoming ? request.requester : request.responder;
	const [showRatingForm, setShowRatingForm] = useState(false);
	const [ratingScore, setRatingScore] = useState(5);
	const [ratingComment, setRatingComment] = useState('');

	// Determine if the *current* user is the requester or responder for this swap
	const isRequester = request.requester_id === currentUserId;
	const myCompleted = isRequester ? request.requester_completed : request.responder_completed;
	const theirCompleted = isRequester ? request.responder_completed : request.requester_completed;

	// Fetch existing ratings (only for completed swaps)
	const { data: swapRatingsData } = useQuery({
		queryKey: ['swapRatings', request.swap_id],
		queryFn: () => api.ratings.getForSwap(request.swap_id),
		enabled: request.status === 'completed',
	});

	const swapRatings: RatingResponse[] = swapRatingsData?.ratings ?? [];
	const myRating = swapRatings.find((r) => r.rater_id === currentUserId);
	const theirRating = swapRatings.find((r) => r.rater_id !== currentUserId);

	const createRatingMutation = useMutation({
		mutationFn: (data: { swap_id: string; ratee_id: string; score: number; comment?: string }) =>
			api.ratings.create(data),
		onSuccess: () => {
			setShowRatingForm(false);
			setRatingScore(5);
			setRatingComment('');
			queryClient.invalidateQueries({ queryKey: ['swapRatings', request.swap_id] });
			toast.success('Rating submitted!');
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to submit rating'),
	});

	const markCompleteMutation = useMutation({
		mutationFn: () => api.swaps.markComplete(request.swap_id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['swaps'] });
			toast.success('Marked as complete!');
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to mark complete'),
	});

	const undoCompleteMutation = useMutation({
		mutationFn: () => api.swaps.undoComplete(request.swap_id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['swaps'] });
			toast.success('Completion undone');
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to undo completion'),
	});

	return (
		<Card className="hover:shadow-md transition-shadow duration-200">
			<CardHeader>
				<div className="flex items-start justify-between">
					<div className="flex items-center space-x-3">
						<Avatar
							src={other.has_photo ? getPhotoUrl(other.user_id) : undefined}
							alt={other.name}
							fallback={
								other.name
									.split(' ')
									.map((n) => n[0])
									.join('') || 'U'
							}
							className="w-12 h-12"
						/>
						<div>
							<h3 className="font-semibold text-lg">{other.name}</h3>
							<p className="text-sm text-muted-foreground">
								{new Date(request.created_at).toLocaleDateString()}
							</p>
						</div>
					</div>
					<div className="flex items-center space-x-2">
						{getStatusIcon(request.status)}
						{getStatusBadge(request.status)}
					</div>
				</div>
			</CardHeader>

			<CardContent className="space-y-4">
				<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div className="p-3 bg-muted rounded-lg">
						<p className="text-sm font-medium text-muted-foreground mb-1">
							{isIncoming ? 'They Offer' : 'You Offer'}
						</p>
						<p className="font-medium">{request.offered_skill.name}</p>
					</div>
					<div className="p-3 bg-muted rounded-lg">
						<p className="text-sm font-medium text-muted-foreground mb-1">
							{isIncoming ? 'They Want' : 'You Want'}
						</p>
						<p className="font-medium">{request.wanted_skill.name}</p>
					</div>
				</div>

				{/* ── Actions ─────────────────────────────────────────── */}
				<div className="flex flex-wrap gap-2 pt-2">
					{/* Pending – incoming: accept / decline */}
					{request.status === 'pending' && isIncoming && (
						<>
							<Button
								onClick={() => onStatusChange(request.swap_id, 'accepted')}
								disabled={statusPending}
								className="flex-1"
							>
								{statusPending ? (
									<Loader2 className="w-4 h-4 mr-2 animate-spin" />
								) : (
									<CheckCircle className="w-4 h-4 mr-2" />
								)}
								Accept
							</Button>
							<Button
								variant="outline"
								onClick={() => onStatusChange(request.swap_id, 'rejected')}
								disabled={statusPending}
								className="flex-1"
							>
								<XCircle className="w-4 h-4 mr-2" />
								Decline
							</Button>
						</>
					)}

					{/* Pending – outgoing: cancel */}
					{request.status === 'pending' && !isIncoming && (
						<Button
							variant="outline"
							onClick={() => onStatusChange(request.swap_id, 'cancelled')}
							disabled={statusPending}
							className="w-full"
						>
							Cancel Request
						</Button>
					)}

					{/* ── Accepted (In Progress) ──────────────────────── */}
					{request.status === 'accepted' && (
						<div className="w-full space-y-3">
							{/* Completion status */}
							<div className="p-3 rounded-lg border border-border bg-muted/50">
								<p className="text-xs font-medium text-muted-foreground mb-2">Completion Status</p>
								<div className="flex items-center gap-3 text-sm">
									<span className={myCompleted ? 'text-green-600 dark:text-green-400 font-medium' : 'text-muted-foreground'}>
										{myCompleted ? '✓ You marked complete' : '○ You: not yet'}
									</span>
									<span className="text-border">|</span>
									<span className={theirCompleted ? 'text-green-600 dark:text-green-400 font-medium' : 'text-muted-foreground'}>
										{theirCompleted ? `✓ ${other.name} marked complete` : `○ ${other.name}: not yet`}
									</span>
								</div>
							</div>

							{/* Action buttons */}
							<div className="flex gap-2">
								<Link href={`/messages?swap=${request.swap_id}`} className="flex-1">
									<Button variant="outline" className="w-full">
										<MessageCircle className="w-4 h-4 mr-2" />
										Open Chat
									</Button>
								</Link>

								{!myCompleted ? (
									<Button
										onClick={() => markCompleteMutation.mutate()}
										disabled={markCompleteMutation.isPending}
										className="flex-1"
									>
										{markCompleteMutation.isPending ? (
											<Loader2 className="w-4 h-4 mr-2 animate-spin" />
										) : (
											<CircleCheckBig className="w-4 h-4 mr-2" />
										)}
										Mark Complete
									</Button>
								) : (
									<Button
										variant="outline"
										onClick={() => undoCompleteMutation.mutate()}
										disabled={undoCompleteMutation.isPending}
										className="flex-1"
									>
										{undoCompleteMutation.isPending ? (
											<Loader2 className="w-4 h-4 mr-2 animate-spin" />
										) : (
											<Undo2 className="w-4 h-4 mr-2" />
										)}
										Undo Complete
									</Button>
								)}
							</div>
						</div>
					)}

					{/* ── Completed – Rating section ──────────────────── */}
					{request.status === 'completed' && (
						<div className="w-full space-y-3">
							{/* Chat link */}
							<Link href={`/messages?swap=${request.swap_id}`}>
								<Button variant="ghost" size="sm" className="text-muted-foreground">
									<MessageCircle className="w-3.5 h-3.5 mr-1.5" />
									View Chat
								</Button>
							</Link>

							{/* Show existing ratings */}
							{myRating && (
								<div className="p-3 bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-900 rounded-lg">
									<div className="flex items-center gap-2">
										<span className="text-xs font-medium text-green-700 dark:text-green-400">You rated</span>
										<div className="flex items-center gap-0.5">
											{[1, 2, 3, 4, 5].map((s) => (
												<Star
													key={s}
													className={`w-3.5 h-3.5 ${
														s <= myRating.score
															? 'fill-yellow-400 text-yellow-400'
															: 'text-gray-300 dark:text-gray-600'
													}`}
												/>
											))}
										</div>
									</div>
									{myRating.comment && (
										<p className="text-sm text-foreground mt-1">{myRating.comment}</p>
									)}
								</div>
							)}
							{theirRating && (
								<div className="p-3 bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900 rounded-lg">
									<div className="flex items-center gap-2">
										<span className="text-xs font-medium text-blue-700 dark:text-blue-400">{other.name}&apos;s rating</span>
										<div className="flex items-center gap-0.5">
											{[1, 2, 3, 4, 5].map((s) => (
												<Star
													key={s}
													className={`w-3.5 h-3.5 ${
														s <= theirRating.score
															? 'fill-yellow-400 text-yellow-400'
															: 'text-gray-300 dark:text-gray-600'
													}`}
												/>
											))}
										</div>
									</div>
									{theirRating.comment && (
										<p className="text-sm text-foreground mt-1">{theirRating.comment}</p>
									)}
								</div>
							)}

							{/* Rating form or button */}
							{!myRating && (
								<>
									{showRatingForm ? (
										<div className="p-3 border border-border rounded-lg space-y-3">
											<div className="flex items-center gap-1">
												{[1, 2, 3, 4, 5].map((s) => (
													<Star
														key={s}
														className={`w-5 h-5 cursor-pointer ${
															s <= ratingScore
																? 'fill-yellow-400 text-yellow-400'
																: 'text-gray-300 dark:text-gray-600'
														}`}
														onClick={() => setRatingScore(s)}
													/>
												))}
											</div>
											<input
												type="text"
												placeholder="Leave a comment (optional)"
												value={ratingComment}
												onChange={(e) => setRatingComment(e.target.value)}
												className="w-full text-sm border border-border rounded-md px-3 py-2 bg-background"
											/>
											<div className="flex gap-2">
												<Button
													size="sm"
													onClick={() =>
														createRatingMutation.mutate({
															swap_id: request.swap_id,
															ratee_id: other.user_id,
															score: ratingScore,
															comment: ratingComment || undefined,
														})
													}
													disabled={createRatingMutation.isPending}
												>
													{createRatingMutation.isPending ? (
														<Loader2 className="w-4 h-4 animate-spin" />
													) : (
														'Submit Rating'
													)}
												</Button>
												<Button
													size="sm"
													variant="outline"
													onClick={() => setShowRatingForm(false)}
												>
													Cancel
												</Button>
											</div>
										</div>
									) : (
										<Button
											variant="outline"
											className="w-full"
											onClick={() => setShowRatingForm(true)}
										>
											<Star className="w-4 h-4 mr-2" />
											Rate Exchange
										</Button>
									)}
								</>
							)}
						</div>
					)}
				</div>
			</CardContent>
		</Card>
	);
}

export default function SwapsPage() {
	return (
		<Suspense fallback={
			<div className="min-h-screen bg-background flex items-center justify-center">
				<Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
			</div>
		}>
			<SwapsContent />
		</Suspense>
	);
}

function SwapsContent() {
	const { user, isLoading: authLoading } = useAuth(true);
	const searchParams = useSearchParams();
	const router = useRouter();
	const requestUserId = searchParams.get('request');
	const queryClient = useQueryClient();
	const [activeTab, setActiveTab] = useState('incoming');

	const { data: swapData, isLoading: swapsLoading } = useQuery({
		queryKey: ['swaps'],
		queryFn: () => api.swaps.list(),
		enabled: !!user,
	});

	const updateStatusMutation = useMutation({
		mutationFn: ({ id, status }: { id: string; status: 'accepted' | 'rejected' | 'cancelled' }) =>
			api.swaps.updateStatus(id, status),
		onSuccess: (_data, variables) => {
			queryClient.invalidateQueries({ queryKey: ['swaps'] });
			if (variables.status === 'accepted') {
				toast.success('Swap accepted! Start chatting to coordinate.');
				router.push(`/messages?swap=${variables.id}`);
			}
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to update swap status'),
	});

	// Swap request creation
	const [showCreateDialog, setShowCreateDialog] = useState(false);
	const [targetUser, setTargetUser] = useState<UserProfileResponse | null>(null);
	const [selectedOffered, setSelectedOffered] = useState<Set<string>>(new Set());
	const [selectedWanted, setSelectedWanted] = useState<Set<string>>(new Set());

	const { data: myProfile } = useQuery({
		queryKey: ['profile'],
		queryFn: () => api.users.getProfile(),
		enabled: !!user && !!requestUserId,
	});

	const { data: targetProfile } = useQuery({
		queryKey: ['publicProfile', requestUserId],
		queryFn: () => api.users.getPublicProfile(requestUserId!),
		enabled: !!requestUserId,
	});

	useEffect(() => {
		if (requestUserId && targetProfile) {
			setTargetUser(targetProfile);
			setShowCreateDialog(true);
		}
	}, [requestUserId, targetProfile]);

	const [swapCreationPending, setSwapCreationPending] = useState(false);

	const handleCreateSwap = async () => {
		if (!targetUser || selectedOffered.size === 0 || selectedWanted.size === 0) return;
		setSwapCreationPending(true);
		try {
			const pairs = Array.from(selectedOffered).flatMap((o) =>
				Array.from(selectedWanted).map((w) => ({ offered: o, wanted: w }))
			);
			for (const pair of pairs) {
				await api.swaps.create({
					responder_id: targetUser.user_id,
					offered_skill_id: pair.offered,
					wanted_skill_id: pair.wanted,
				});
			}
			setShowCreateDialog(false);
			setTargetUser(null);
			setSelectedOffered(new Set());
			setSelectedWanted(new Set());
			router.replace('/swaps');
			queryClient.invalidateQueries({ queryKey: ['swaps'] });
			toast.success(`Created ${pairs.length} swap request${pairs.length > 1 ? 's' : ''}`);
		} catch (err: unknown) {
			toast.error(err instanceof Error ? err.message : 'Failed to create swap request');
		} finally {
			setSwapCreationPending(false);
		}
	};

	const closeCreateDialog = () => {
		setShowCreateDialog(false);
		setTargetUser(null);
		setSelectedOffered(new Set());
		setSelectedWanted(new Set());
		router.replace('/swaps');
	};

	const toggleOffered = (id: string) => {
		setSelectedOffered((prev) => {
			const next = new Set(prev);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			return next;
		});
	};

	const toggleWanted = (id: string) => {
		setSelectedWanted((prev) => {
			const next = new Set(prev);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			return next;
		});
	};

	if (authLoading || swapsLoading) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
			</div>
		);
	}

	const incoming = swapData?.received ?? [];
	const outgoing = swapData?.sent ?? [];
	const allSwaps = [...incoming, ...outgoing];

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Header */}
				<div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between mb-10 animate-fade-in-up">
					<div>
						<h1 className="text-display-md text-foreground mb-2">Swap Requests</h1>
						<p className="text-muted-foreground text-lg">Manage your skill exchange requests</p>
					</div>
					<Link href="/browse" className="self-start sm:self-auto">
						<Button>
							<ArrowRight className="w-4 h-4 mr-2" />
							Find People
						</Button>
					</Link>
				</div>

				{/* Create Swap Request Dialog */}
				{showCreateDialog && targetUser && (
					<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
						<Card className="w-full max-w-md mx-4">
							<CardHeader>
								<div className="flex items-center justify-between">
									<h2 className="text-lg font-semibold">Request Skill Swap</h2>
									<Button variant="ghost" size="sm" onClick={closeCreateDialog}>
										<X className="w-4 h-4" />
									</Button>
								</div>
							</CardHeader>
							<CardContent className="space-y-4">
								{/* Target user info */}
								<div className="flex items-center gap-3 p-3 bg-muted rounded-lg">
									<Avatar
										src={targetUser.has_photo ? getPhotoUrl(targetUser.user_id) : undefined}
										alt={targetUser.name}
										fallback={targetUser.name
											.split(' ')
											.map((n) => n[0])
											.join('')}
										className="w-10 h-10"
									/>
									<div>
										<p className="font-medium">{targetUser.name}</p>
										{targetUser.location && (
											<p className="text-sm text-muted-foreground">{targetUser.location}</p>
										)}
									</div>
								</div>

								{/* Skills you can teach them */}
								<div>
									<label className="text-sm font-medium mb-2 block">
										Skills you offer{' '}
										<span className="text-muted-foreground font-normal">(matching what they want)</span>
									</label>
									<div className="flex flex-wrap gap-2">
										{(myProfile?.skills_offered ?? [])
											.filter((skill: SkillResponse) => {
												const theirWantedIds = new Set(
													(targetUser.skills_wanted ?? []).map((s: SkillResponse) => s.skill_id)
												);
												return theirWantedIds.has(skill.skill_id);
											})
											.map((skill: SkillResponse) => (
												<button
													key={skill.skill_id}
													type="button"
													onClick={() => toggleOffered(skill.skill_id)}
													className={`inline-flex items-center rounded-full px-3 py-1.5 text-sm font-medium border transition-colors cursor-pointer ${
														selectedOffered.has(skill.skill_id)
															? 'bg-primary text-primary-foreground border-primary'
															: 'bg-background text-foreground border-border hover:bg-muted'
													}`}
												>
													{skill.name}
													{selectedOffered.has(skill.skill_id) && (
														<CheckCircle className="w-3.5 h-3.5 ml-1.5" />
													)}
												</button>
											))}
										{(myProfile?.skills_offered ?? []).filter((skill: SkillResponse) => {
											const theirWantedIds = new Set(
												(targetUser.skills_wanted ?? []).map((s: SkillResponse) => s.skill_id)
											);
											return theirWantedIds.has(skill.skill_id);
										}).length === 0 && (
											<p className="text-sm text-muted-foreground">No matching skills found</p>
										)}
									</div>
								</div>

								{/* Skills you want from them */}
								<div>
									<label className="text-sm font-medium mb-2 block">
										Skills you want from them
									</label>
									<div className="flex flex-wrap gap-2">
										{(targetUser.skills_offered ?? []).map((skill: SkillResponse) => (
											<button
												key={skill.skill_id}
												type="button"
												onClick={() => toggleWanted(skill.skill_id)}
												className={`inline-flex items-center rounded-full px-3 py-1.5 text-sm font-medium border transition-colors cursor-pointer ${
													selectedWanted.has(skill.skill_id)
														? 'bg-primary text-primary-foreground border-primary'
														: 'bg-background text-foreground border-border hover:bg-muted'
												}`}
											>
												{skill.name}
												{selectedWanted.has(skill.skill_id) && (
													<CheckCircle className="w-3.5 h-3.5 ml-1.5" />
												)}
											</button>
										))}
									</div>
								</div>

								<div className="flex gap-2 pt-2">
									<Button
										className="flex-1"
										onClick={handleCreateSwap}
										disabled={selectedOffered.size === 0 || selectedWanted.size === 0 || swapCreationPending}
									>
										{swapCreationPending ? (
											<Loader2 className="w-4 h-4 mr-2 animate-spin" />
										) : (
											<ArrowRightLeft className="w-4 h-4 mr-2" />
										)}
										{selectedOffered.size > 0 && selectedWanted.size > 0
											? `Send ${selectedOffered.size * selectedWanted.size} Request${selectedOffered.size * selectedWanted.size > 1 ? 's' : ''}`
											: 'Send Request'}
									</Button>
									<Button variant="outline" onClick={closeCreateDialog}>
										Cancel
									</Button>
								</div>
							</CardContent>
						</Card>
					</div>
				)}

				{/* Stats */}
				<div className="grid grid-cols-2 sm:grid-cols-4 gap-px mb-10 rounded-xl bg-border border border-border overflow-hidden animate-fade-in">
					<div className="flex items-center gap-2 bg-muted/40 px-4 py-4">
						<Clock className="w-4 h-4 text-yellow-500 shrink-0" />
						<span className="text-2xl font-bold tabular-nums">{allSwaps.filter((r) => r.status === 'pending').length}</span>
						<span className="text-sm text-muted-foreground">Pending</span>
					</div>
					<div className="flex items-center gap-2 bg-muted/40 px-4 py-4">
						<MessageCircle className="w-4 h-4 text-blue-500 shrink-0" />
						<span className="text-2xl font-bold tabular-nums">{allSwaps.filter((r) => r.status === 'accepted').length}</span>
						<span className="text-sm text-muted-foreground">In Progress</span>
					</div>
					<div className="flex items-center gap-2 bg-muted/40 px-4 py-4">
						<CircleCheckBig className="w-4 h-4 text-green-500 shrink-0" />
						<span className="text-2xl font-bold tabular-nums">{allSwaps.filter((r) => r.status === 'completed').length}</span>
						<span className="text-sm text-muted-foreground">Completed</span>
					</div>
					<div className="flex items-center gap-2 bg-muted/40 px-4 py-4">
						<User className="w-4 h-4 text-purple-500 shrink-0" />
						<span className="text-2xl font-bold tabular-nums">{allSwaps.length}</span>
						<span className="text-sm text-muted-foreground">Total</span>
					</div>
				</div>

				{/* Tabs */}
				<Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
					<TabsList className="grid w-full grid-cols-2">
						<TabsTrigger value="incoming">
							Incoming ({incoming.length})
						</TabsTrigger>
						<TabsTrigger value="outgoing">
							Outgoing ({outgoing.length})
						</TabsTrigger>
					</TabsList>

					<TabsContent value="incoming" className="space-y-4">
						{incoming.length > 0 ? (
							incoming.map((request) => (
								<SwapCard
									key={request.swap_id}
									request={request}
									isIncoming
									currentUserId={user!.user_id}
									onStatusChange={(id, status) => updateStatusMutation.mutate({ id, status })}
									statusPending={updateStatusMutation.isPending}
								/>
							))
						) : (
							<div className="text-center py-12 text-muted-foreground">
								<p>No incoming swap requests yet</p>
							</div>
						)}
					</TabsContent>

					<TabsContent value="outgoing" className="space-y-4">
						{outgoing.length > 0 ? (
							outgoing.map((request) => (
								<SwapCard
									key={request.swap_id}
									request={request}
									isIncoming={false}
									currentUserId={user!.user_id}
									onStatusChange={(id, status) => updateStatusMutation.mutate({ id, status })}
									statusPending={updateStatusMutation.isPending}
								/>
							))
						) : (
							<div className="text-center py-12 text-muted-foreground">
								<p>
									No outgoing swap requests yet.{' '}
									<Link href="/browse" className="text-primary underline">
										Browse users
									</Link>{' '}
									to start!
								</p>
							</div>
						)}
					</TabsContent>
				</Tabs>
			</div>
		</div>
	);
}
