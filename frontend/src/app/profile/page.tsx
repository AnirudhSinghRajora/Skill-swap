'use client';

import { useState, useRef, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import imageCompression from 'browser-image-compression';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar } from '@/components/ui/avatar';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
	Star,
	MapPin,
	Calendar,
	Edit,
	Save,
	X,
	CheckCircle,
	XCircle,
	Clock,
	Loader2,
	Plus,
	Camera,
} from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api, { getPhotoUrl, updateStoredUser } from '@/lib/api';
import { toast } from 'sonner';

export default function ProfilePage() {
	return (
		<Suspense>
			<ProfileContent />
		</Suspense>
	);
}

function ProfileContent() {
	const { user, isLoading: authLoading } = useAuth(true);
	const queryClient = useQueryClient();
	const searchParams = useSearchParams();
	const [activeTab, setActiveTab] = useState(searchParams.get('tab') || 'overview');
	const [isEditing, setIsEditing] = useState(false);
	const [editForm, setEditForm] = useState({ name: '', email: '', location: '', is_public: true });
	const [managingOffered, setManagingOffered] = useState(false);
	const [managingWanted, setManagingWanted] = useState(false);
	const [pendingOffered, setPendingOffered] = useState<Set<string>>(new Set());
	const [pendingWanted, setPendingWanted] = useState<Set<string>>(new Set());
	const [savingOffered, setSavingOffered] = useState(false);
	const [savingWanted, setSavingWanted] = useState(false);
	const photoInputRef = useRef<HTMLInputElement>(null);

	const { data: profile, isLoading: profileLoading } = useQuery({
		queryKey: ['profile'],
		queryFn: () => api.users.getProfile(),
		enabled: !!user,
	});

	const { data: swapData } = useQuery({
		queryKey: ['swaps'],
		queryFn: () => api.swaps.list(),
		enabled: !!user,
	});

	const { data: ratingStats } = useQuery({
		queryKey: ['rating-stats', user?.user_id],
		queryFn: () => api.ratings.getStatsForUser(user!.user_id),
		enabled: !!user?.user_id,
	});

	const { data: userRatings } = useQuery({
		queryKey: ['ratings', user?.user_id],
		queryFn: () => api.ratings.getForUser(user!.user_id, { limit: 10 }),
		enabled: !!user?.user_id,
	});

	const { data: allSkills } = useQuery({
		queryKey: ['skills'],
		queryFn: () => api.skills.list(),
		enabled: !!user,
	});

	const updateProfileMutation = useMutation({
		mutationFn: (data: { name?: string; email?: string; location?: string; is_public?: boolean }) =>
			api.users.updateProfile(data),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['profile'] });
			setIsEditing(false);
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to update profile'),
	});

	const removeOfferedMutation = useMutation({
		mutationFn: (skillId: string) => api.skills.removeOffered(skillId),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: ['profile'] }),
		onError: (err: Error) => toast.error(err.message || 'Failed to remove skill'),
	});

	const removeWantedMutation = useMutation({
		mutationFn: (skillId: string) => api.skills.removeWanted(skillId),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: ['profile'] }),
		onError: (err: Error) => toast.error(err.message || 'Failed to remove skill'),
	});

	const togglePending = (set: Set<string>, setFn: (s: Set<string>) => void, id: string) => {
		const next = new Set(set);
		if (next.has(id)) next.delete(id); else next.add(id);
		setFn(next);
	};

	const saveOffered = async () => {
		if (pendingOffered.size === 0) { setManagingOffered(false); return; }
		setSavingOffered(true);
		try {
			await Promise.all([...pendingOffered].map((id) => api.skills.addOffered(id)));
			queryClient.invalidateQueries({ queryKey: ['profile'] });
			setPendingOffered(new Set());
		} catch (err: unknown) {
			toast.error(err instanceof Error ? err.message : 'Failed to add skills');
		} finally {
			setSavingOffered(false);
			setManagingOffered(false);
		}
	};

	const saveWanted = async () => {
		if (pendingWanted.size === 0) { setManagingWanted(false); return; }
		setSavingWanted(true);
		try {
			await Promise.all([...pendingWanted].map((id) => api.skills.addWanted(id)));
			queryClient.invalidateQueries({ queryKey: ['profile'] });
			setPendingWanted(new Set());
		} catch (err: unknown) {
			toast.error(err instanceof Error ? err.message : 'Failed to add skills');
		} finally {
			setSavingWanted(false);
			setManagingWanted(false);
		}
	};

	const uploadPhotoMutation = useMutation({
		mutationFn: (file: File) => api.files.uploadPhoto(file),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['profile'] });
			updateStoredUser({ has_photo: true });
			toast.success('Profile photo updated!');
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to upload photo'),
	});

	const handlePhotoChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
		const file = e.target.files?.[0];
		if (file) {
			if (file.size > 10 * 1024 * 1024) {
				toast.error('Photo must be under 10MB');
				return;
			}
			try {
				const compressed = await imageCompression(file, {
					maxSizeMB: 1,
					maxWidthOrHeight: 800,
					useWebWorker: true,
				});
				uploadPhotoMutation.mutate(new File([compressed], file.name, { type: compressed.type }));
			} catch {
				toast.error('Failed to compress photo');
			}
		}
	};

	if (authLoading || profileLoading) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
			</div>
		);
	}

	if (!profile) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<div className="text-center">
					<h1 className="text-2xl font-bold text-foreground mb-2">Profile not found</h1>
					<p className="text-muted-foreground">Unable to load your profile.</p>
				</div>
			</div>
		);
	}

	const allSwaps = [...(swapData?.sent ?? []), ...(swapData?.received ?? [])];
	const recentSwaps = allSwaps
		.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
		.slice(0, 5);

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

	const getStatusColor = (status: string) => {
		switch (status) {
			case 'accepted':
				return 'text-green-600';
			case 'rejected':
				return 'text-red-600';
			case 'pending':
				return 'text-yellow-600';
			default:
				return 'text-muted-foreground';
		}
	};

	const startEditing = () => {
		setEditForm({
			name: profile.name,
			email: profile.email,
			location: profile.location ?? '',
			is_public: profile.is_public,
		});
		setIsEditing(true);
	};

	const handleSaveProfile = () => {
		updateProfileMutation.mutate({
			name: editForm.name,
			email: editForm.email,
			location: editForm.location,
			is_public: editForm.is_public,
		});
	};

	// Filter out skills user already has
	const availableOfferedSkills =
		allSkills?.filter(
			(s) => !profile.skills_offered.some((o) => o.skill_id === s.skill_id)
		) ?? [];
	const availableWantedSkills =
		allSkills?.filter(
			(s) => !profile.skills_wanted.some((w) => w.skill_id === s.skill_id)
		) ?? [];

	const pendingOfferedSkills = availableOfferedSkills.filter((s) => pendingOffered.has(s.skill_id));
	const unpickedOfferedSkills = availableOfferedSkills.filter((s) => !pendingOffered.has(s.skill_id));
	const pendingWantedSkills = availableWantedSkills.filter((s) => pendingWanted.has(s.skill_id));
	const unpickedWantedSkills = availableWantedSkills.filter((s) => !pendingWanted.has(s.skill_id));

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Profile Header */}
				<Card className="mb-8 overflow-hidden">
					<CardContent className="p-6 sm:p-8">
						<div className="flex flex-col md:flex-row items-start md:items-center gap-6">
						<div className="relative group">
							<Avatar
							src={profile.has_photo ? getPhotoUrl(profile.user_id) : undefined}
								alt={profile.name}
								fallback={profile.name
									.split(' ')
									.map((n) => n[0])
									.join('')}
								className="w-24 h-24"
							/>
							<button
								type="button"
								onClick={() => photoInputRef.current?.click()}
								className="absolute inset-0 flex items-center justify-center bg-black/50 rounded-full opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
							>
								{uploadPhotoMutation.isPending ? (
									<Loader2 className="w-6 h-6 text-white animate-spin" />
								) : (
									<Camera className="w-6 h-6 text-white" />
								)}
							</button>
							<input
								ref={photoInputRef}
								type="file"
								accept="image/*"
								className="hidden"
								onChange={handlePhotoChange}
							/>
						</div>

						<div className="flex-1">
							{isEditing ? (
								<div className="space-y-3">
									<Input
										value={editForm.name}
										onChange={(e) =>
											setEditForm((f) => ({ ...f, name: e.target.value }))
										}
										placeholder="Full Name"
									/>
									<Input
										type="email"
										value={editForm.email}
										onChange={(e) =>
											setEditForm((f) => ({ ...f, email: e.target.value }))
										}
										placeholder="Email"
									/>
									<Input
										value={editForm.location}
										onChange={(e) =>
											setEditForm((f) => ({ ...f, location: e.target.value }))
										}
										placeholder="Location"
									/>
									<div className="flex items-center gap-3">
										<Button
											type="button"
											size="sm"
											variant={editForm.is_public ? 'default' : 'outline'}
											onClick={() =>
												setEditForm((f) => ({ ...f, is_public: !f.is_public }))
											}
										>
											{editForm.is_public ? 'Public' : 'Private'}
										</Button>
										<span className="text-sm text-muted-foreground">
											{editForm.is_public
												? 'Others can find you in search'
												: 'Your profile is hidden from search'}
										</span>
									</div>
									<div className="flex gap-2">
										<Button
											size="sm"
											onClick={handleSaveProfile}
											disabled={updateProfileMutation.isPending}
										>
											{updateProfileMutation.isPending ? (
												<Loader2 className="w-4 h-4 mr-2 animate-spin" />
											) : (
												<Save className="w-4 h-4 mr-2" />
											)}
											Save
										</Button>
										<Button
											size="sm"
											variant="outline"
											onClick={() => setIsEditing(false)}
										>
											<X className="w-4 h-4 mr-2" />
											Cancel
										</Button>
									</div>
								</div>
							) : (
								<div className="flex flex-col md:flex-row md:items-center md:justify-between">
									<div>
										<h1 className="text-display-md text-foreground mb-2">
											{profile.name}
										</h1>
										<div className="flex items-center space-x-4 text-muted-foreground mb-3">
											{profile.location && (
												<div className="flex items-center space-x-1">
													<MapPin className="w-4 h-4" />
													<span>{profile.location}</span>
												</div>
											)}
											<div className="flex items-center space-x-1">
												<Calendar className="w-4 h-4" />
												<span>
													Joined{' '}
													{new Date(profile.created_at).toLocaleDateString()}
												</span>
											</div>
										</div>
										<div className="flex items-center space-x-2">
											<Star className="w-5 h-5 fill-yellow-400 text-yellow-400" />
											<span className="font-medium">
											{ratingStats?.average_rating
												? ratingStats.average_rating.toFixed(1)
													: 'No rating'}
											</span>
											<span className="text-muted-foreground">
												({ratingStats?.total_ratings ?? 0}{' '}
												{(ratingStats?.total_ratings ?? 0) === 1
													? 'review'
													: 'reviews'})
											</span>
										</div>
									</div>

									<div className="mt-4 md:mt-0">
										<Button variant="outline" onClick={startEditing}>
											<Edit className="w-4 h-4 mr-2" />
											Edit Profile
										</Button>
									</div>
								</div>
							)}
						</div>
						</div>
					</CardContent>
				</Card>

				{/* Tabs */}
				<Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
					<TabsList className="inline-flex w-full sm:w-auto">
						<TabsTrigger value="overview">Overview</TabsTrigger>
						<TabsTrigger value="skills">Skills</TabsTrigger>
						<TabsTrigger value="swaps">Swap History</TabsTrigger>
						<TabsTrigger value="reviews">Reviews</TabsTrigger>
					</TabsList>

					<TabsContent value="overview" className="space-y-6">
						<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
							<Card>
								<CardHeader>
									<CardTitle>Skills Offered</CardTitle>
								</CardHeader>
								<CardContent>
									<div className="flex flex-wrap gap-2">
										{profile.skills_offered.length > 0 ? (
											profile.skills_offered.map((skill) => (
												<Badge key={skill.skill_id} variant="secondary">
													{skill.name}
												</Badge>
											))
										) : (
											<p className="text-sm text-muted-foreground">
												No skills offered yet
											</p>
										)}
									</div>
								</CardContent>
							</Card>

							<Card>
								<CardHeader>
									<CardTitle>Skills Wanted</CardTitle>
								</CardHeader>
								<CardContent>
									<div className="flex flex-wrap gap-2">
										{profile.skills_wanted.length > 0 ? (
											profile.skills_wanted.map((skill) => (
												<Badge key={skill.skill_id} variant="outline">
													{skill.name}
												</Badge>
											))
										) : (
											<p className="text-sm text-muted-foreground">
												No skills wanted yet
											</p>
										)}
									</div>
								</CardContent>
							</Card>
						</div>

						<Card>
							<CardHeader>
								<CardTitle>Recent Activity</CardTitle>
							</CardHeader>
							<CardContent>
								{recentSwaps.length > 0 ? (
									<div className="space-y-4">
										{recentSwaps.map((request) => {
											const isRequester = request.requester_id === user?.user_id;
											const other = isRequester
												? request.responder
												: request.requester;

											return (
												<div
													key={request.swap_id}
													className="flex items-center justify-between p-3 border border-border rounded-lg"
												>
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
															className="w-10 h-10"
														/>
														<div>
															<p className="font-medium">
																{isRequester
																	? 'You offered'
																	: `${other.name} offered`}{' '}
																{request.offered_skill.name}
															</p>
															<p className="text-sm text-muted-foreground">
																for {request.wanted_skill.name}
															</p>
														</div>
													</div>
													<div className="flex items-center space-x-2">
														{getStatusIcon(request.status)}
														<span
															className={`text-sm font-medium ${getStatusColor(request.status)}`}
														>
															{request.status.charAt(0).toUpperCase() +
																request.status.slice(1)}
														</span>
													</div>
												</div>
											);
										})}
									</div>
								) : (
									<p className="text-center text-muted-foreground py-4">
										No swap activity yet
									</p>
								)}
							</CardContent>
						</Card>
					</TabsContent>

					<TabsContent value="skills" className="space-y-6">
						<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
							<Card>
								<CardHeader className="flex flex-row items-center justify-between space-y-0">
									<CardTitle>Skills I Can Teach</CardTitle>
									<Button
										size="sm"
										variant={managingOffered ? 'default' : 'outline'}
										onClick={() => managingOffered ? saveOffered() : setManagingOffered(true)}
										disabled={savingOffered}
									>
										{savingOffered ? (
											<><Loader2 className="w-4 h-4 mr-1 animate-spin" /> Saving...</>
										) : managingOffered ? (
											<><CheckCircle className="w-4 h-4 mr-1" /> Done{pendingOffered.size > 0 && ` (${pendingOffered.size})`}</>
										) : (
											<><Edit className="w-4 h-4 mr-1" /> Manage</>
										)}
									</Button>
								</CardHeader>
								<CardContent>
									{/* Current skills */}
									<div className="flex flex-wrap gap-2">
										{profile.skills_offered.map((skill) => (
											<span
												key={skill.skill_id}
												className="inline-flex items-center rounded-full px-3 py-1.5 text-sm font-medium bg-primary text-primary-foreground"
											>
												{skill.name}
												{managingOffered && (
													<button
														type="button"
														onClick={() => removeOfferedMutation.mutate(skill.skill_id)}
														className="ml-1.5 hover:opacity-70 cursor-pointer"
														disabled={removeOfferedMutation.isPending}
													>
														<X className="w-3.5 h-3.5" />
													</button>
												)}
											</span>
										))}
										{/* Pending additions (not yet saved) */}
										{pendingOfferedSkills.map((skill) => (
											<button
												key={skill.skill_id}
												type="button"
												onClick={() => togglePending(pendingOffered, setPendingOffered, skill.skill_id)}
												className="inline-flex items-center rounded-full px-3 py-1.5 text-sm font-medium bg-primary/20 text-primary border border-primary border-dashed cursor-pointer transition-colors"
											>
												{skill.name}
												<X className="w-3.5 h-3.5 ml-1.5" />
											</button>
										))}
										{profile.skills_offered.length === 0 && pendingOffered.size === 0 && !managingOffered && (
											<p className="text-sm text-muted-foreground py-4 w-full text-center">
												No skills offered yet. Click Manage to add some!
											</p>
										)}
									</div>
									{/* Available skills to add */}
									{managingOffered && unpickedOfferedSkills.length > 0 && (
										<div className="mt-4 pt-4 border-t border-border">
											<p className="text-xs text-muted-foreground mb-2">Click to add:</p>
											<div className="flex flex-wrap gap-2">
												{unpickedOfferedSkills.map((s) => (
													<button
														key={s.skill_id}
														type="button"
														onClick={() => togglePending(pendingOffered, setPendingOffered, s.skill_id)}
														className="inline-flex items-center rounded-full px-3 py-1.5 text-sm font-medium border border-dashed border-border bg-background text-foreground hover:bg-muted transition-colors cursor-pointer"
													>
														<Plus className="w-3.5 h-3.5 mr-1" />
														{s.name}
													</button>
												))}
											</div>
										</div>
									)}
								</CardContent>
							</Card>

							<Card>
								<CardHeader className="flex flex-row items-center justify-between space-y-0">
									<CardTitle>Skills I Want to Learn</CardTitle>
									<Button
										size="sm"
										variant={managingWanted ? 'default' : 'outline'}
										onClick={() => managingWanted ? saveWanted() : setManagingWanted(true)}
										disabled={savingWanted}
									>
										{savingWanted ? (
											<><Loader2 className="w-4 h-4 mr-1 animate-spin" /> Saving...</>
										) : managingWanted ? (
											<><CheckCircle className="w-4 h-4 mr-1" /> Done{pendingWanted.size > 0 && ` (${pendingWanted.size})`}</>
										) : (
											<><Edit className="w-4 h-4 mr-1" /> Manage</>
										)}
									</Button>
								</CardHeader>
								<CardContent>
									{/* Current skills */}
									<div className="flex flex-wrap gap-2">
										{profile.skills_wanted.map((skill) => (
											<span
												key={skill.skill_id}
												className="inline-flex items-center rounded-full px-3 py-1.5 text-sm font-medium bg-primary text-primary-foreground"
											>
												{skill.name}
												{managingWanted && (
													<button
														type="button"
														onClick={() => removeWantedMutation.mutate(skill.skill_id)}
														className="ml-1.5 hover:opacity-70 cursor-pointer"
														disabled={removeWantedMutation.isPending}
													>
														<X className="w-3.5 h-3.5" />
													</button>
												)}
											</span>
										))}
										{/* Pending additions (not yet saved) */}
										{pendingWantedSkills.map((skill) => (
											<button
												key={skill.skill_id}
												type="button"
												onClick={() => togglePending(pendingWanted, setPendingWanted, skill.skill_id)}
												className="inline-flex items-center rounded-full px-3 py-1.5 text-sm font-medium bg-primary/20 text-primary border border-primary border-dashed cursor-pointer transition-colors"
											>
												{skill.name}
												<X className="w-3.5 h-3.5 ml-1.5" />
											</button>
										))}
										{profile.skills_wanted.length === 0 && pendingWanted.size === 0 && !managingWanted && (
											<p className="text-sm text-muted-foreground py-4 w-full text-center">
												No skills wanted yet. Click Manage to add some!
											</p>
										)}
									</div>
									{/* Available skills to add */}
									{managingWanted && unpickedWantedSkills.length > 0 && (
										<div className="mt-4 pt-4 border-t border-border">
											<p className="text-xs text-muted-foreground mb-2">Click to add:</p>
											<div className="flex flex-wrap gap-2">
												{unpickedWantedSkills.map((s) => (
													<button
														key={s.skill_id}
														type="button"
														onClick={() => togglePending(pendingWanted, setPendingWanted, s.skill_id)}
														className="inline-flex items-center rounded-full px-3 py-1.5 text-sm font-medium border border-dashed border-border bg-background text-foreground hover:bg-muted transition-colors cursor-pointer"
													>
														<Plus className="w-3.5 h-3.5 mr-1" />
														{s.name}
													</button>
												))}
											</div>
										</div>
									)}
								</CardContent>
							</Card>
						</div>
					</TabsContent>

					<TabsContent value="swaps" className="space-y-6">
						<Card>
							<CardHeader>
								<CardTitle>Swap History</CardTitle>
							</CardHeader>
							<CardContent>
								{allSwaps.length > 0 ? (
									<div className="space-y-4">
										{allSwaps.map((request) => {
											const isRequester = request.requester_id === user?.user_id;
											const other = isRequester
												? request.responder
												: request.requester;

											return (
												<div
													key={request.swap_id}
													className="flex items-center justify-between p-4 border border-border rounded-lg"
												>
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
															className="w-10 h-10"
														/>
														<div>
															<p className="font-medium">
																{isRequester ? 'You' : other.name} offered{' '}
																{request.offered_skill.name}
															</p>
															<p className="text-sm text-muted-foreground">
																for {request.wanted_skill.name} •{' '}
																{new Date(
																	request.created_at
																).toLocaleDateString()}
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
															{request.status.charAt(0).toUpperCase() +
																request.status.slice(1)}
														</Badge>
													</div>
												</div>
											);
										})}
									</div>
								) : (
									<p className="text-center text-muted-foreground py-8">
										No swap history yet. Browse users to start exchanging skills!
									</p>
								)}
							</CardContent>
						</Card>
					</TabsContent>

					<TabsContent value="reviews" className="space-y-6">
						<Card>
							<CardHeader>
								<CardTitle>
									Reviews ({ratingStats?.total_ratings ?? 0})
									{ratingStats?.average_rating && (
										<span className="ml-2 text-sm font-normal text-muted-foreground">
											Average: {ratingStats.average_rating.toFixed(1)} / 5
										</span>
									)}
								</CardTitle>
							</CardHeader>
							<CardContent>
								{userRatings && userRatings.length > 0 ? (
									<div className="space-y-4">
										{userRatings.map((rating) => (
											<div
												key={rating.rating_id}
												className="p-4 border border-border rounded-lg"
											>
												{rating.rater && (
													<div className="mb-2">
														<Link
															href={`/profile/${rating.rater.user_id}`}
															className="text-sm font-medium text-primary hover:underline"
														>
															{rating.rater.name}
														</Link>
													</div>
												)}
												<div className="flex items-center justify-between mb-2">
													<div className="flex items-center space-x-1">
														{Array.from({ length: 5 }).map((_, i) => (
															<Star
																key={i}
																className={`w-4 h-4 ${
																	i < rating.score
																		? 'fill-yellow-400 text-yellow-400'
																		: 'text-gray-300 dark:text-gray-600'
																}`}
															/>
														))}
													</div>
													<span className="text-sm text-muted-foreground">
														{new Date(
															rating.created_at
														).toLocaleDateString()}
													</span>
												</div>
												{rating.comment ? (
													<p className="text-sm text-foreground">
														{rating.comment}
													</p>
												) : (
													<p className="text-sm text-muted-foreground italic">
														Rated {rating.score} star{rating.score !== 1 ? 's' : ''}
													</p>
												)}
											</div>
										))}
									</div>
								) : (
									<p className="text-center text-muted-foreground py-8">
										No reviews yet
									</p>
								)}
							</CardContent>
						</Card>
					</TabsContent>
				</Tabs>
			</div>
		</div>
	);
}
