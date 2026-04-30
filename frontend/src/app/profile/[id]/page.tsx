'use client';

import { useState, use } from 'react';
import { useSearchParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar } from '@/components/ui/avatar';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
	MapPin,
	Star,
	Clock,
	ArrowRightLeft,
	Loader2,
	ArrowLeft,
	Calendar,
} from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useQuery } from '@tanstack/react-query';
import api, { getPhotoUrl } from '@/lib/api';
import { getDayLabels, formatTime } from '@/types/availability';
import Link from 'next/link';

export default function PublicProfilePage({
	params,
}: {
	params: Promise<{ id: string }>;
}) {
	const { id: userId } = use(params);
	const { user } = useAuth();
	const searchParams = useSearchParams();
	const initialTab = searchParams.get('tab') === 'reviews' ? 'reviews' : 'skills';

	const { data: profile, isLoading: profileLoading, error: profileError } = useQuery({
		queryKey: ['publicProfile', userId],
		queryFn: () => api.users.getPublicProfile(userId),
	});

	const { data: ratings = [] } = useQuery({
		queryKey: ['userRatings', userId],
		queryFn: () => api.ratings.getForUser(userId, { limit: 10 }),
		enabled: !!profile,
	});

	const { data: ratingStats } = useQuery({
		queryKey: ['userRatingStats', userId],
		queryFn: () => api.ratings.getStatsForUser(userId),
		enabled: !!profile,
	});

	const { data: commonSlots = [] } = useQuery({
		queryKey: ['commonAvailability', userId],
		queryFn: () => api.availability.findCommon(userId),
		enabled: !!profile && !!user,
	});

	const [activeTab, setActiveTab] = useState(initialTab);

	if (profileLoading) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
			</div>
		);
	}

	if (profileError || !profile) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<div className="text-center space-y-4">
					<h1 className="text-2xl font-bold text-foreground">User Not Found</h1>
					<p className="text-muted-foreground">
						This profile may be private or the user does not exist.
					</p>
					<Link href="/browse">
						<Button variant="outline">
							<ArrowLeft className="w-4 h-4 mr-2" />
							Back to Browse
						</Button>
					</Link>
				</div>
			</div>
		);
	}

	const isOwnProfile = user?.user_id === profile.user_id;

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Back button */}
				<Link href="/browse" className="inline-flex items-center text-muted-foreground hover:text-foreground mb-6">
					<ArrowLeft className="w-4 h-4 mr-1" />
					Back to Browse
				</Link>

				{/* Profile header */}
				<Card className="mb-6">
					<CardContent className="p-6">
						<div className="flex flex-col sm:flex-row items-start gap-6">
							<Avatar
								src={profile.has_photo ? getPhotoUrl(profile.user_id) : undefined}
								alt={profile.name}
								fallback={profile.name
									.split(' ')
									.map((n) => n[0])
									.join('')}
								className="w-24 h-24"
							/>
							<div className="flex-1 space-y-2">
								<h1 className="text-2xl font-bold text-foreground">{profile.name}</h1>
								{profile.location && (
									<div className="flex items-center gap-1 text-muted-foreground">
										<MapPin className="w-4 h-4" />
										<span>{profile.location}</span>
									</div>
								)}
								<div className="flex items-center gap-4">
									{ratingStats && ratingStats.total_ratings > 0 && (
										<div className="flex items-center gap-1">
											<Star className="w-4 h-4 fill-yellow-400 text-yellow-400" />
											<span className="font-medium">
												{ratingStats.average_rating.toFixed(1)}
											</span>
											<span className="text-muted-foreground text-sm">
												({ratingStats.total_ratings} review{ratingStats.total_ratings !== 1 ? 's' : ''})
											</span>
										</div>
									)}
									<div className="flex items-center gap-1 text-muted-foreground text-sm">
										<Calendar className="w-4 h-4" />
										<span>Joined {new Date(profile.created_at).toLocaleDateString()}</span>
									</div>
								</div>
							</div>
							{!isOwnProfile && (
								<Link href={`/swaps?request=${profile.user_id}`}>
									<Button>
										<ArrowRightLeft className="w-4 h-4 mr-2" />
										Request Swap
									</Button>
								</Link>
							)}
							{isOwnProfile && (
								<Link href="/profile">
									<Button variant="outline">Edit Profile</Button>
								</Link>
							)}
						</div>
					</CardContent>
				</Card>

				{/* Tabs */}
				<Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
					<TabsList className="grid w-full grid-cols-3">
						<TabsTrigger value="skills">Skills</TabsTrigger>
						<TabsTrigger value="reviews">Reviews</TabsTrigger>
						<TabsTrigger value="availability">Availability</TabsTrigger>
					</TabsList>

					{/* Skills Tab */}
					<TabsContent value="skills" className="space-y-6">
						<Card>
							<CardHeader>
								<CardTitle>Skills Offered</CardTitle>
							</CardHeader>
							<CardContent>
								{profile.skills_offered.length > 0 ? (
									<div className="flex flex-wrap gap-2">
										{profile.skills_offered.map((skill) => (
											<Badge key={skill.skill_id} variant="secondary" className="text-sm px-3 py-1">
												{skill.name}
											</Badge>
										))}
									</div>
								) : (
									<p className="text-muted-foreground">No skills offered yet.</p>
								)}
							</CardContent>
						</Card>
						<Card>
							<CardHeader>
								<CardTitle>Skills Wanted</CardTitle>
							</CardHeader>
							<CardContent>
								{profile.skills_wanted.length > 0 ? (
									<div className="flex flex-wrap gap-2">
										{profile.skills_wanted.map((skill) => (
											<Badge key={skill.skill_id} variant="outline" className="text-sm px-3 py-1">
												{skill.name}
											</Badge>
										))}
									</div>
								) : (
									<p className="text-muted-foreground">No skills wanted yet.</p>
								)}
							</CardContent>
						</Card>
					</TabsContent>

					{/* Reviews Tab */}
					<TabsContent value="reviews" className="space-y-4">
						{ratingStats && ratingStats.total_ratings > 0 && (
							<div className="flex items-center gap-3 mb-2">
								<div className="flex items-center gap-1">
									<Star className="w-5 h-5 fill-yellow-400 text-yellow-400" />
									<span className="text-lg font-semibold">{ratingStats.average_rating.toFixed(1)}</span>
								</div>
								<span className="text-muted-foreground text-sm">
									{ratingStats.total_ratings} review{ratingStats.total_ratings !== 1 ? 's' : ''}
								</span>
							</div>
						)}
						{ratings.length > 0 ? (
							ratings.map((rating) => (
								<Card key={rating.rating_id}>
									<CardContent className="p-4">
										<div className="flex items-center justify-between mb-2">
											<div className="flex items-center gap-2">
												{rating.rater && (
													<Link
														href={`/profile/${rating.rater.user_id}`}
														className="text-sm font-medium hover:underline"
													>
														{rating.rater.name}
													</Link>
												)}
												<div className="flex items-center gap-0.5">
													{[1, 2, 3, 4, 5].map((s) => (
														<Star
															key={s}
															className={`w-3.5 h-3.5 ${
																s <= rating.score
																	? 'fill-yellow-400 text-yellow-400'
																: 'text-gray-300 dark:text-gray-600'
														}`}
													/>
												))}
											</div>
										</div>
										<span className="text-xs text-muted-foreground">
											{new Date(rating.created_at).toLocaleDateString()}
										</span>
									</div>
									{rating.comment ? (
										<p className="text-sm text-foreground">{rating.comment}</p>
									) : (
										<p className="text-sm text-muted-foreground italic">Rated {rating.score} star{rating.score !== 1 ? 's' : ''}</p>
										)}
									</CardContent>
								</Card>
							))
						) : (
							<Card>
								<CardContent className="p-8 text-center text-muted-foreground">
									No reviews yet.
								</CardContent>
							</Card>
						)}
					</TabsContent>

					{/* Availability Tab */}
					<TabsContent value="availability" className="space-y-4">
						{user && !isOwnProfile && commonSlots.length > 0 && (
							<Card className="border-primary/50">
								<CardHeader>
									<CardTitle className="text-lg flex items-center gap-2">
										<Clock className="w-5 h-5 text-primary" />
										Common Availability
									</CardTitle>
								</CardHeader>
								<CardContent className="space-y-3">
									{commonSlots.map((slot) => (
										<div key={slot.slot_id} className="p-3 bg-primary/5 rounded-lg">
											<p className="font-medium text-sm">{slot.label}</p>
											<div className="flex flex-wrap gap-1 mt-1">
												{getDayLabels(slot.day_bitmask).map((day) => (
													<span
														key={day}
														className="text-xs px-2 py-0.5 bg-primary/10 text-primary rounded-full"
													>
														{day}
													</span>
												))}
											</div>
											<p className="text-sm text-muted-foreground mt-1">
												{formatTime(slot.start_time)} – {formatTime(slot.end_time)}
											</p>
										</div>
									))}
								</CardContent>
							</Card>
						)}

						{user && !isOwnProfile && commonSlots.length === 0 && (
							<Card>
								<CardContent className="p-8 text-center text-muted-foreground">
									No common availability found. Set your own availability in your{' '}
									<Link href="/profile" className="text-primary underline">
										profile
									</Link>{' '}
									to find matching times.
								</CardContent>
							</Card>
						)}

						{!user && (
							<Card>
								<CardContent className="p-8 text-center text-muted-foreground">
									<Link href="/auth/signin" className="text-primary underline">
										Sign in
									</Link>{' '}
									to see common availability.
								</CardContent>
							</Card>
						)}
					</TabsContent>
				</Tabs>
			</div>
		</div>
	);
}
