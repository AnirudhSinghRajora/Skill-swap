'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Avatar } from '@/components/ui/avatar';
import { Bell, Home, Users, User, Settings, LogOut, ArrowRight, Menu, X, MessageCircle } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useUnreadCount } from '@/hooks/useUnreadCount';
import { useQuery } from '@tanstack/react-query';
import api, { getPhotoUrl } from '@/lib/api';

export function Navigation() {
	const pathname = usePathname();
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
	const { isAuthenticated, isLoading: authLoading, user, logout } = useAuth();

	const { data: notifStats } = useQuery({
		queryKey: ['notification-stats'],
		queryFn: () => api.notifications.stats(),
		enabled: isAuthenticated,
		refetchInterval: 30000, // poll every 30s
	});

	const unreadCount = notifStats?.unread_count ?? 0;
	const chatUnread = useUnreadCount();

	const navItems = [
		{ href: '/dashboard', label: 'Dashboard', icon: Home },
		{ href: '/browse', label: 'Browse', icon: Users },
		{ href: '/messages', label: 'Messages', icon: MessageCircle, badge: chatUnread },
		{ href: '/swaps', label: 'Swaps', icon: ArrowRight },
		{ href: '/profile', label: 'Profile', icon: User },
		{ href: '/settings', label: 'Settings', icon: Settings }
	];

	// Don't show nav on auth pages
	const isAuthPage = pathname?.startsWith('/auth');
	if (isAuthPage) return null;

	return (
		<nav className="sticky top-0 z-50 bg-card border-b border-border">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="flex justify-between items-center h-16">
					{/* Logo */}
					<Link href="/" className="flex items-center space-x-2">
						<div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
							<span className="text-primary-foreground font-bold text-sm">SS</span>
						</div>
						<span className="font-bold text-xl text-foreground">SkillSwap</span>
					</Link>

					{/* Navigation Links - only show when authenticated */}
					{isAuthenticated && (
						<div className="hidden md:flex items-center space-x-1">
							{navItems.map((item) => {
								const Icon = item.icon;
								const isActive = pathname === item.href || (item.href !== '/dashboard' && pathname?.startsWith(item.href));
								const badgeCount = 'badge' in item ? (item.badge as number) : 0;

								return (
									<Link key={item.href} href={item.href}>
										<Button
											variant={isActive ? 'default' : 'ghost'}
											size="sm"
											className="relative flex items-center space-x-2"
										>
											<Icon className="w-4 h-4" />
											<span>{item.label}</span>
											{badgeCount > 0 && (
												<span className="absolute -top-1 -right-1 bg-destructive text-destructive-foreground text-[10px] rounded-full w-5 h-5 flex items-center justify-center">
													{badgeCount > 99 ? '99+' : badgeCount}
												</span>
											)}
										</Button>
									</Link>
								);
							})}
						</div>
					)}

					{/* Right side */}
					<div className="flex items-center space-x-4">
						{isAuthenticated ? (
							<>
								{/* Notifications */}
								<Link href="/notifications">
									<Button variant="ghost" size="sm" className="relative">
										<Bell className="w-5 h-5" />
										{unreadCount > 0 && (
											<span className="absolute -top-1 -right-1 bg-destructive text-destructive-foreground text-xs rounded-full w-5 h-5 flex items-center justify-center">
												{unreadCount > 99 ? '99+' : unreadCount}
											</span>
										)}
									</Button>
								</Link>

								{/* Profile Avatar */}
								<Link href="/profile">
									<Avatar
										src={user?.has_photo ? getPhotoUrl(user.user_id) : undefined}
										alt={user?.name || 'Profile'}
										fallback={user?.name?.charAt(0)?.toUpperCase() || 'U'}
										className="w-8 h-8 cursor-pointer"
									/>
								</Link>

								{/* Logout */}
								<Button variant="ghost" size="sm" onClick={logout}>
									<LogOut className="w-4 h-4" />
								</Button>

								{/* Mobile menu toggle */}
								<Button
									variant="ghost"
									size="sm"
									className="md:hidden"
									onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
								>
									{mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
								</Button>
							</>
						) : authLoading ? (
							<div className="w-20" /> /* placeholder to prevent layout shift */
						) : (
							<div className="flex items-center space-x-2">
								<Link href="/auth/signin">
									<Button variant="ghost" size="sm">Sign In</Button>
								</Link>
								<Link href="/auth/signup">
									<Button size="sm">Sign Up</Button>
								</Link>
							</div>
						)}
					</div>
				</div>

				{/* Mobile Navigation Menu */}
				{isAuthenticated && mobileMenuOpen && (
					<div className="md:hidden border-t border-border py-2 space-y-1">
						{navItems.map((item) => {
							const Icon = item.icon;
							const isActive = pathname === item.href || (item.href !== '/dashboard' && pathname?.startsWith(item.href));
							const badgeCount = 'badge' in item ? (item.badge as number) : 0;

							return (
								<Link key={item.href} href={item.href} onClick={() => setMobileMenuOpen(false)}>
									<Button
										variant={isActive ? 'default' : 'ghost'}
										size="sm"
										className="w-full justify-start flex items-center space-x-2"
									>
										<Icon className="w-4 h-4" />
										<span>{item.label}</span>
										{badgeCount > 0 && (
											<span className="ml-auto bg-destructive text-destructive-foreground text-[10px] rounded-full w-5 h-5 flex items-center justify-center">
												{badgeCount > 99 ? '99+' : badgeCount}
											</span>
										)}
									</Button>
								</Link>
							);
						})}
					</div>
				)}
			</div>
		</nav>
	);
}
