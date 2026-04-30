'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useTheme } from 'next-themes';
import { Button } from '@/components/ui/button';
import { Avatar } from '@/components/ui/avatar';
import {
	Bell,
	Home,
	Users,
	User,
	LogOut,
	ArrowRight,
	Menu,
	X,
	MessageCircle,
	Moon,
	Sun
} from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useUnreadCount } from '@/hooks/useUnreadCount';
import { useQuery } from '@tanstack/react-query';
import api, { getPhotoUrl } from '@/lib/api';

/**
 * Magazine-masthead nav.
 *  Row 1 — folio bar: issue number · date · live-swappers wire (citron pulse).
 *           Collapses to 0px on scroll to make the nav compact.
 *  Row 2 — main bar: serif wordmark · sweep-underline nav · glyph cluster.
 */
export function Navigation() {
	const pathname = usePathname();
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
	const [scrolled, setScrolled] = useState(false);
	const { isAuthenticated, isLoading: authLoading, user, logout } = useAuth();
	const { theme, setTheme } = useTheme();

	const { data: notifStats } = useQuery({
		queryKey: ['notification-stats'],
		queryFn: () => api.notifications.stats(),
		enabled: isAuthenticated,
		refetchInterval: 30000
	});

	const unreadCount = notifStats?.unread_count ?? 0;
	const chatUnread = useUnreadCount();

	useEffect(() => {
		const onScroll = () => setScrolled(window.scrollY > 12);
		onScroll();
		window.addEventListener('scroll', onScroll, { passive: true });
		return () => window.removeEventListener('scroll', onScroll);
	}, []);

	const marketingItems = [
		{ href: '/browse', label: 'Browse' },
		{ href: '/#how', label: 'Method' },
		{ href: '/#features', label: 'Features' }
	];

	const appItems = [
		{ href: '/dashboard', label: 'Dashboard', icon: Home },
		{ href: '/browse', label: 'Browse', icon: Users },
		{ href: '/messages', label: 'Messages', icon: MessageCircle, badge: chatUnread },
		{ href: '/swaps', label: 'Swaps', icon: ArrowRight },
		{ href: '/profile', label: 'Profile', icon: User }
	];

	const isAuthPage = pathname?.startsWith('/auth');
	if (isAuthPage) return null;

	const today = new Intl.DateTimeFormat('en-GB', {
		day: '2-digit',
		month: 'short',
		year: 'numeric'
	}).format(new Date()).toUpperCase();

	const issueNo = String(
		Math.max(1, Math.floor((Date.now() - Date.parse('2024-01-01')) / 86_400_000))
	).padStart(3, '0');

	return (
		<nav className="sticky top-0 z-50 bg-background/85 backdrop-blur-lg">
			{/* ── Row 1 — folio bar ───────────────────────────────────── */}
			<div
				aria-hidden="true"
				className={[
					'overflow-hidden border-b border-foreground/10 transition-[max-height,opacity] duration-300 ease-out',
					scrolled ? 'max-h-0 opacity-0' : 'max-h-8 opacity-100'
				].join(' ')}
			>
				<div className="max-w-[1400px] mx-auto px-6 sm:px-10">
					<div className="flex items-center justify-between h-8 text-[10px] tracking-[0.22em] uppercase text-muted-foreground font-mono">
						<span className="hidden sm:inline-block">
							Vol. 01 · Issue&nbsp;
							<span className="text-foreground tabular-nums">№{issueNo}</span>
						</span>
						<span className="hidden md:inline-block tabular-nums">{today}</span>
						<span className="flex items-center gap-2">
							<span className="relative flex h-1.5 w-1.5">
								<span className="absolute inset-0 rounded-full bg-accent animate-ping opacity-70" />
								<span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-accent" />
							</span>
							<span className="text-foreground/80">Swaps live now</span>
						</span>
					</div>
				</div>
			</div>

			{/* ── Row 2 — main bar ────────────────────────────────────── */}
			<div
				className={[
					'border-b border-foreground/10 transition-[height] duration-300 ease-out',
					scrolled ? 'shadow-[0_1px_0_0_color-mix(in_oklch,var(--foreground)_8%,transparent)]' : ''
				].join(' ')}
			>
				<div className="max-w-[1400px] mx-auto px-6 sm:px-10">
					<div className={`flex items-center justify-between transition-[height] duration-300 ease-out ${scrolled ? 'h-14' : 'h-16'}`}>
						{/* Wordmark — serif lockup with terminal slash */}
						<Link href="/" className="group flex items-center gap-2">
							<span className="flex items-baseline">
								<span className="font-serif text-[1.65rem] leading-none tracking-tight text-foreground">
									Skill
								</span>
								<span className="font-serif text-[1.65rem] italic leading-none tracking-tight text-primary transition-transform duration-200 group-hover:translate-x-px">
									swap
								</span>
								<span className="font-serif text-[1.4rem] leading-none text-foreground/40">.</span>
							</span>
							<span
								className="hidden lg:inline-block text-[9px] tracking-[0.28em] uppercase text-muted-foreground font-mono pl-2 ml-1 border-l border-foreground/15"
								aria-hidden="true"
							>
								An exchange
							</span>
						</Link>

						{/* Center — sweep-underline nav */}
						{isAuthenticated ? (
							<div className="hidden lg:flex items-center">
								{appItems.map((item) => {
									const Icon = item.icon;
									const isActive =
										pathname === item.href ||
										(item.href !== '/dashboard' && pathname?.startsWith(item.href));
									const badgeCount = 'badge' in item ? (item.badge as number) : 0;
									return (
										<Link
											key={item.href}
											href={item.href}
											className={[
												'group relative inline-flex items-center gap-1.5 px-3 py-2 text-[13px] tracking-tight transition-colors',
												isActive ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
											].join(' ')}
										>
											<Icon className="w-3.5 h-3.5" />
											<span>{item.label}</span>
											{badgeCount > 0 && (
												<span className="ml-1 inline-flex items-center justify-center rounded-full bg-foreground text-background text-[10px] tabular-nums px-1.5 min-w-[18px] h-[18px]">
													{badgeCount > 99 ? '99+' : badgeCount}
												</span>
											)}
											{/* sweep underline */}
											<span
												className={[
													'absolute left-3 right-3 -bottom-px h-px origin-left transition-transform duration-300 ease-out',
													isActive ? 'bg-primary scale-x-100' : 'bg-accent scale-x-0 group-hover:scale-x-100'
												].join(' ')}
											/>
										</Link>
									);
								})}
							</div>
						) : (
							<div className="hidden md:flex items-center">
								{marketingItems.map((item) => (
									<Link
										key={item.href}
										href={item.href}
										className="group relative px-3 py-2 text-[13px] tracking-tight text-muted-foreground hover:text-foreground transition-colors"
									>
										<span>{item.label}</span>
										<span className="absolute left-3 right-3 -bottom-px h-px origin-left bg-accent scale-x-0 group-hover:scale-x-100 transition-transform duration-300 ease-out" />
									</Link>
								))}
							</div>
						)}

						{/* Right cluster */}
						<div className="flex items-center gap-1">
							<button
								onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
								className="relative w-8 h-8 inline-flex items-center justify-center rounded-full text-muted-foreground hover:text-foreground hover:bg-foreground/5 transition-colors"
								aria-label="Toggle theme"
							>
								<Sun className="w-4 h-4 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" />
								<Moon className="absolute w-4 h-4 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" />
							</button>

							{isAuthenticated ? (
								<>
									<Link
										href="/notifications"
										className="relative w-8 h-8 inline-flex items-center justify-center rounded-full text-muted-foreground hover:text-foreground hover:bg-foreground/5 transition-colors"
										aria-label="Notifications"
									>
										<Bell className="w-4 h-4" />
										{unreadCount > 0 && (
											<span className="absolute top-0.5 right-0.5 inline-flex items-center justify-center rounded-full bg-accent text-accent-foreground text-[9px] font-medium tabular-nums min-w-[14px] h-[14px] px-1 animate-badge-pulse">
												{unreadCount > 99 ? '99+' : unreadCount}
											</span>
										)}
									</Link>

									<Link href="/profile" className="ml-1">
										<Avatar
											src={user?.has_photo ? getPhotoUrl(user.user_id) : undefined}
											alt={user?.name || 'Profile'}
											fallback={user?.name?.charAt(0)?.toUpperCase() || 'U'}
											className="w-7 h-7 cursor-pointer ring-1 ring-foreground/10 hover:ring-primary/40 transition"
										/>
									</Link>

									<button
										onClick={logout}
										className="w-8 h-8 inline-flex items-center justify-center rounded-full text-muted-foreground hover:text-foreground hover:bg-foreground/5 transition-colors"
										aria-label="Log out"
									>
										<LogOut className="w-3.5 h-3.5" />
									</button>

									<button
										className="lg:hidden w-8 h-8 inline-flex items-center justify-center rounded-full text-muted-foreground hover:text-foreground hover:bg-foreground/5 transition-colors"
										onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
										aria-label="Menu"
									>
										{mobileMenuOpen ? <X className="w-4 h-4" /> : <Menu className="w-4 h-4" />}
									</button>
								</>
							) : authLoading ? (
								<div className="w-24" />
							) : (
								<div className="flex items-center gap-1.5 ml-1">
									<Link href="/auth/signin">
										<Button variant="ghost" size="sm" className="text-[13px]">
											Sign in
										</Button>
									</Link>
									<Link href="/auth/signup" className="hidden sm:block">
										<Button variant="brand" size="sm" className="text-[13px] gap-1.5">
											Start a swap
											<ArrowRight className="w-3.5 h-3.5" />
										</Button>
									</Link>
								</div>
							)}
						</div>
					</div>

					{/* Mobile drawer */}
					{isAuthenticated && mobileMenuOpen && (
						<div className="lg:hidden border-t border-foreground/10 py-2 space-y-0.5">
							{appItems.map((item) => {
								const Icon = item.icon;
								const isActive =
									pathname === item.href ||
									(item.href !== '/dashboard' && pathname?.startsWith(item.href));
								const badgeCount = 'badge' in item ? (item.badge as number) : 0;
								return (
									<Link
										key={item.href}
										href={item.href}
										onClick={() => setMobileMenuOpen(false)}
										className={[
											'flex items-center gap-2 px-3 py-2.5 rounded-md text-sm transition-colors',
											isActive
												? 'bg-foreground/5 text-foreground'
												: 'text-muted-foreground hover:bg-foreground/5 hover:text-foreground'
										].join(' ')}
									>
										<Icon className="w-4 h-4" />
										<span>{item.label}</span>
										{badgeCount > 0 && (
											<span className="ml-auto inline-flex items-center justify-center rounded-full bg-foreground text-background text-[10px] tabular-nums px-1.5 min-w-[18px] h-[18px]">
												{badgeCount > 99 ? '99+' : badgeCount}
											</span>
										)}
									</Link>
								);
							})}
						</div>
					)}
				</div>
			</div>
		</nav>
	);
}
