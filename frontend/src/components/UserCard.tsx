'use client';

import { UserType } from '@/types/user';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar } from '@/components/ui/avatar';
import { Star, MapPin, ArrowRightLeft, ArrowRight } from 'lucide-react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import api, { getPhotoUrl } from '@/lib/api';

interface UserCardProps {
	user: UserType;
	currentUserId?: string;
}

export function UserCard({ user, currentUserId }: UserCardProps) {
	const isCurrentUser = currentUserId === user.user_id;

	const { data: ratingStats } = useQuery({
		queryKey: ['userRatingStats', user.user_id],
		queryFn: () => api.ratings.getStatsForUser(user.user_id),
	});

	return (
		<Card className="hover:shadow-md transition-shadow duration-200">
			<CardHeader className="pb-3">
				<div className="flex items-start justify-between">
					<div className="flex items-center space-x-3">
						<Avatar
							src={user.has_photo ? getPhotoUrl(user.user_id) : undefined}
							alt={user.name}
							fallback={user.name
								.split(' ')
								.map((n) => n[0])
								.join('')}
							className="w-12 h-12"
						/>
						<div>
							<h3 className="font-semibold text-lg">{user.name}</h3>
							<div className="flex items-center space-x-2 text-sm text-muted-foreground">
								{user.location && (
									<>
										<MapPin className="w-4 h-4" />
										<span>{user.location}</span>
									</>
								)}
							</div>
						</div>
					</div>
					<Link
						href={`/profile/${user.user_id}?tab=reviews`}
						className="flex items-center space-x-1 hover:opacity-80 transition-opacity"
						title="View reviews"
					>
						<Star className="w-4 h-4 fill-yellow-400 text-yellow-400" />
						{ratingStats && ratingStats.total_ratings > 0 ? (
							<>
								<span className="text-sm font-medium">{ratingStats.average_rating.toFixed(1)}</span>
								<span className="text-xs text-muted-foreground">({ratingStats.total_ratings})</span>
							</>
						) : (
							<span className="text-sm font-medium text-muted-foreground">New</span>
						)}
					</Link>
				</div>
			</CardHeader>

			<CardContent className="space-y-4">
				{/* Skills Offered */}
				<div>
					<h4 className="text-sm font-medium text-muted-foreground mb-2">
						Skills Offered
					</h4>
					<div className="flex flex-wrap gap-1">
						{user.skills_offered
							.filter((skill) => skill)
							.map((skill) => (
								<Badge key={skill.skill_id} variant="secondary" className="text-xs">
									{skill.name}
								</Badge>
							))}
					</div>
				</div>

				{/* Skills Wanted */}
				<div>
					<h4 className="text-sm font-medium text-muted-foreground mb-2">
						Skills Wanted
					</h4>
					<div className="flex flex-wrap gap-1">
						{user.skills_wanted
							.filter((skill) => skill)
							.map((skill) => (
								<Badge key={skill.skill_id} variant="outline" className="text-xs">
									{skill.name}
								</Badge>
							))}
					</div>
				</div>

				{/* Actions */}
				<div className="flex space-x-2 pt-2">
					{!isCurrentUser ? (
						<>
							<Link href={`/swaps?request=${user.user_id}`}>
								<Button variant="outline" size="sm" className="flex-1">
									<ArrowRightLeft className="w-4 h-4 mr-2" />
									Request Swap
								</Button>
							</Link>
							<Link href={`/profile/${user.user_id}`}>
								<Button size="sm" className="flex-1">
									<ArrowRight className="w-4 h-4 mr-2" />
									View Profile
								</Button>
							</Link>
						</>
					) : (
						<Button variant="outline" size="sm" className="w-full" disabled>
							This is you
						</Button>
					)}
				</div>
			</CardContent>
		</Card>
	);
}
