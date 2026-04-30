'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { UserCard } from '@/components/UserCard';
import { Search, Filter, MapPin } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useQuery } from '@tanstack/react-query';
import api from '@/lib/api';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent } from '@/components/ui/card';

export default function BrowsePage() {
	const { user } = useAuth(true);
	const [searchTerm, setSearchTerm] = useState('');
	const [locationFilter, setLocationFilter] = useState('');
	const [selectedSkills, setSelectedSkills] = useState<string[]>([]);
	const [page, setPage] = useState(1);
	const limit = 12;

	const { data: searchResults, isLoading: usersLoading } = useQuery({
		queryKey: ['browse-users', searchTerm, locationFilter, page],
		queryFn: () =>
			api.users.searchPublic({
				search_term: searchTerm || undefined,
				location: locationFilter || undefined,
				page,
				limit,
			}),
		enabled: !!user,
	});

	const { data: allSkills } = useQuery({
		queryKey: ['skills'],
		queryFn: () => api.skills.list(),
		enabled: !!user,
	});

	const users = searchResults?.users ?? [];
	const totalPages = searchResults?.total_pages ?? 1;
	const total = searchResults?.total ?? 0;

	// Client-side skill filtering (search API doesn't support skill filtering)
	const filteredUsers =
		selectedSkills.length === 0
			? users
			: users.filter((u) =>
					selectedSkills.some(
						(skillId) =>
							u.skills_offered.some((s) => s.skill_id === skillId) ||
							u.skills_wanted.some((s) => s.skill_id === skillId)
					)
				);

	const toggleSkillFilter = (skillId: string) => {
		setSelectedSkills((prev) =>
			prev.includes(skillId) ? prev.filter((id) => id !== skillId) : [...prev, skillId]
		);
	};

	const clearFilters = () => {
		setSearchTerm('');
		setSelectedSkills([]);
		setLocationFilter('');
		setPage(1);
	};

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Header */}
				<div className="mb-10 animate-fade-in-up flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
					<div>
						<h1 className="text-display-md text-foreground mb-2">Discover People</h1>
						<p className="text-muted-foreground text-lg">
							Find people to exchange skills with in your area
						</p>
					</div>
					{(searchTerm || selectedSkills.length > 0 || locationFilter) && (
						<Button
							variant="outline"
							size="sm"
							onClick={clearFilters}
							className="self-start md:self-auto"
						>
							Clear filters
						</Button>
					)}
				</div>

				{/* Search and Filters */}
				<div className="mb-10 space-y-5">
					{/* Top row: search + location, side by side on md+ */}
					<div className="grid gap-3 sm:grid-cols-[1fr_auto]">
						<div className="relative">
							<Search className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground w-4 h-4" />
							<Input
								placeholder="Search by name or skill..."
								value={searchTerm}
								onChange={(e) => {
									setSearchTerm(e.target.value);
									setPage(1);
								}}
								className="pl-10"
							/>
						</div>
						<div className="relative sm:w-64">
							<MapPin className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground w-4 h-4" />
							<Input
								placeholder="Location"
								value={locationFilter}
								onChange={(e) => {
									setLocationFilter(e.target.value);
									setPage(1);
								}}
								className="pl-10"
							/>
						</div>
					</div>

					{/* Skill chips row */}
					{allSkills && allSkills.length > 0 && (
						<div className="flex flex-wrap items-center gap-2">
							<div className="flex items-center gap-1.5 text-xs uppercase tracking-[0.14em] text-muted-foreground mr-1">
								<Filter className="w-3.5 h-3.5" />
								<span>Skills</span>
							</div>
							{allSkills.slice(0, 12).map((skill) => (
								<Badge
									key={skill.skill_id}
									variant={selectedSkills.includes(skill.skill_id) ? 'default' : 'outline'}
									className="cursor-pointer hover:bg-primary/10 transition-colors"
									onClick={() => toggleSkillFilter(skill.skill_id)}
								>
									{skill.name}
								</Badge>
							))}
						</div>
					)}
				</div>

				{/* Results count */}
				<div className="mb-6">
					<div className="flex items-center justify-between">
						<h2 className="text-xl font-semibold">
							{filteredUsers.length}{' '}
							{filteredUsers.length === 1 ? 'person' : 'people'} found
						</h2>
						{total > 0 && (
							<p className="text-sm text-muted-foreground">
								Page {page} of {totalPages} ({total} total)
							</p>
						)}
					</div>
				</div>

				{/* Loading */}
				{usersLoading && (
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
						{[1, 2, 3, 4, 5, 6].map((i) => (
							<Card key={i}>
								<CardContent className="p-6">
									<div className="flex items-center gap-4 mb-4">
										<Skeleton className="h-12 w-12 rounded-full" />
										<div>
											<Skeleton className="h-5 w-28 mb-1" />
											<Skeleton className="h-3 w-20" />
										</div>
									</div>
									<div className="space-y-2">
										<Skeleton className="h-4 w-full" />
										<div className="flex gap-2">
											<Skeleton className="h-6 w-16 rounded-full" />
											<Skeleton className="h-6 w-20 rounded-full" />
											<Skeleton className="h-6 w-14 rounded-full" />
										</div>
									</div>
								</CardContent>
							</Card>
						))}
					</div>
				)}

				{/* User Grid */}
				{!usersLoading && filteredUsers.length > 0 && (
					<div className="grid grid-cols-[repeat(auto-fill,minmax(320px,1fr))] gap-6">
						{filteredUsers.map((u) => (
							<UserCard key={u.user_id} user={u} currentUserId={user?.user_id} />
						))}
					</div>
				)}

				{/* Empty State */}
				{!usersLoading && filteredUsers.length === 0 && (
					<div className="text-center py-16">
						<div className="w-16 h-16 mx-auto mb-5 bg-muted rounded-2xl flex items-center justify-center">
							<Search className="w-8 h-8 text-muted-foreground" />
						</div>
						<h3 className="text-lg font-semibold text-foreground mb-2">No matches found</h3>
						<p className="text-muted-foreground mb-6 max-w-sm mx-auto">
							Try broadening your search or adjusting filters to discover more people
						</p>
						<Button variant="outline" onClick={clearFilters}>Clear all filters</Button>
					</div>
				)}

				{/* Pagination */}
				{totalPages > 1 && (
					<div className="flex items-center justify-center gap-2 mt-8">
						<Button
							variant="outline"
							size="sm"
							disabled={page <= 1}
							onClick={() => setPage((p) => p - 1)}
						>
							Previous
						</Button>
						{Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
							const pageNum = i + 1;
							return (
								<Button
									key={pageNum}
									variant={page === pageNum ? 'default' : 'outline'}
									size="sm"
									onClick={() => setPage(pageNum)}
								>
									{pageNum}
								</Button>
							);
						})}
						<Button
							variant="outline"
							size="sm"
							disabled={page >= totalPages}
							onClick={() => setPage((p) => p + 1)}
						>
							Next
						</Button>
					</div>
				)}
			</div>
		</div>
	);
}
