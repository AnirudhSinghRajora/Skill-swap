'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useTheme } from 'next-themes';
import { Button } from '@/components/ui/button';
import { Avatar } from '@/components/ui/avatar';
import { Bell, Home, Users, User, Settings, LogOut, ArrowRight, Menu, X, MessageCircle, Moon, Sun } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useUnreadCount } from '@/hooks/useUnreadCount';
import { useQuery } from '@tanstack/react-query';
import api, { getPhotoUrl } from '@/lib/api';

export function Navigation() {
	const pathname = usePathname();
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
	const { isAuthenticated, isLoading: authLoading, user, logout } = useAuth();
	const { theme, setTheme } = useTheme();

	const { data: notifStats } = useQuery({
		queryKey: ['notification-stats'],
		queryFn: () => api.notifications.stats(),
		enabled: isAuthenticated,
		refetchInterval: 30000,
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

	const isAuthPage = pathname?.startsWith('/auth');
	if (isAuthPage) return null;

	return (
		<nav className="sticky top-0 z-50 bg-background/85 backdrop-blur-lg border-b border-foreground/10">
			<div className="max-w-[1400px] mx-auto px-6 sm:px-10">
				<div className="flex justify-between items-center h-16">
					{/* Wordmark — editorial serif lockup */}
					<Link href="/" className="group flex items-baseline gap-1">
						<span className="font-serif text-2xl leading-none tracking-tight text-foreground">
							Skill
						</span>
						<span className="font-serif text-2xl italic leading-none tracking-tight text-primary transition-transform duration-200 group-hover:translate-x-px">
							swap
						</span>
						<span className="font-serif text-xl leading-none text-foreground/45">.</span>
					</Link>

					{/* Navigation Links */}
					{isAuthenticated && (
						<div className="hidden md:flex items-center gap-1">
							{navItems.map((item) => {
								const Icon = item.icon;
								const isActive = pathname === item.href || (item.href !== '/dashboard' && pathname?.startsWith(item.href));
								const badgeCount = 'badge' in item ? (item.badge as number) : 0;

								return (
									<Link key={item.href} href={item.href}>
										<Button
											variant={isActive ? 'default' : 'ghost'}
											size="sm"
											className="relative gap-1.5 text-sm"
										>
											<Icon className="w-3.5 h-3.5" />
											<span>{item.label}</span>
											{badgeCount > 0 && (
												<span className="absolute -top-1 -right-1 bg-destructive text-destructive-foreground text-[10px] rounded-full w-4.5 h-4.5 flex items-center justify-center">
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
					<div className="flex items-center gap-1">
						{/* Theme toggle */}
						<Button
							variant="ghost"
							size="sm"
							onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
							className="w-8 h-8 p-0"
						>
							<Sun className="w-4 h-4 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" />
							<Moon className="absolute w-4 h-4 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" />
							<span className="sr-only">Toggle theme</span>
						</Button>

						{isAuthenticated ? (
							<>
								{/* Notifications */}
								<Link href="/notifications">
									<Button variant="ghost" size="sm" className="relative w-8 h-8 p-0">
										<Bell className="w-4 h-4" />
										{unreadCount > 0 && (
											<span className="absolute -top-0.5 -right-0.5 bg-destructive text-destructive-foreground text-[10px] rounded-full w-4 h-4 flex items-center justify-center animate-badge-pulse">
												{unreadCount > 99 ? '99+' : unreadCount}
											</span>
										)}
									</Button>
								</Link>

								{/* Profile Avatar */}
								<Link href="/profile" className="ml-1">
									<Avatar
										src={user?.has_photo ? getPhotoUrl(user.user_id) : undefined}
										alt={user?.name || 'Profile'}
										fallback={user?.name?.charAt(0)?.toUpperCase() || 'U'}
										className="w-7 h-7 cursor-pointer"
									/>
								</Link>

								{/* Logout */}
								<Button variant="ghost" size="sm" onClick={logout} className="w-8 h-8 p-0">
									<LogOut className="w-3.5 h-3.5" />
								</Button>

								{/* Mobile menu toggle */}
								<Button
									variant="ghost"
									size="sm"
									className="md:hidden w-8 h-8 p-0"
									onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
								>
									{mobileMenuOpen ? <X className="w-4 h-4" /> : <Menu className="w-4 h-4" />}
								</Button>
							</>
						) : authLoading ? (
							<div className="w-20" />
						) : (
							<div className="flex items-center gap-2">
								<Link href="/auth/signin">
									<Button variant="ghost" size="sm" className="text-sm">Sign in</Button>
								</Link>
								<Link href="/auth/signup">
									<Button variant="brand" size="sm" className="text-sm">Start a swap</Button>
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
										className="w-full justify-start gap-2"
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
