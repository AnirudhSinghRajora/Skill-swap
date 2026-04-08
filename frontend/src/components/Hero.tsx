'use client';

import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { ArrowRight, CalendarDays, MessageCircleMore, Sparkles, Video } from 'lucide-react';
import Link from 'next/link';

const HeroSection = () => {
	return (
		<section className="relative flex items-center justify-center overflow-hidden bg-background">
			<div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_15%_25%,color-mix(in_oklab,var(--primary)_14%,transparent)_0%,transparent_46%),radial-gradient(circle_at_82%_16%,color-mix(in_oklab,var(--secondary)_42%,transparent)_0%,transparent_44%),linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:auto,auto,4rem_4rem,4rem_4rem] [mask-image:radial-gradient(ellipse_78%_68%_at_50%_22%,#000_58%,transparent_100%)] opacity-90" />

			<div className="relative z-10 mx-auto w-full max-w-6xl px-4 pt-28 pb-20 md:px-6 md:pt-32 md:pb-24">
				{/* Pill badge */}
				<motion.div
					initial={{ opacity: 0, y: 16 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.5 }}
					className="mb-8 inline-flex items-center gap-2 rounded-full border border-border bg-card/80 px-4 py-1.5 text-sm text-muted-foreground backdrop-blur"
				>
					<span className="w-1.5 h-1.5 rounded-full bg-primary" />
					Peer-to-peer skill exchange platform
				</motion.div>

				{/* Headline */}
				<motion.h1
					initial={{ opacity: 0, y: 20 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.6, delay: 0.1 }}
					className="mb-6 max-w-4xl text-balance text-display-xl text-foreground"
				>
					Learn anything.
					<br />
					<span className="text-muted-foreground">Teach everything.</span>
				</motion.h1>

				{/* Subheading */}
				<motion.p
					initial={{ opacity: 0, y: 20 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.6, delay: 0.2 }}
					className="mb-10 max-w-3xl text-pretty text-lg leading-relaxed text-muted-foreground md:text-xl"
				>
					Exchange your expertise directly with others. No middlemen, no fees -
					just real people trading real skills.
				</motion.p>

				{/* CTAs */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.6, delay: 0.3 }}
					className="mb-12 flex flex-col items-start gap-3 sm:flex-row sm:items-center"
				>
					<Link href="/auth/signup">
						<Button size="lg" className="px-8 h-11 text-sm font-medium group">
							Get Started
							<ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
						</Button>
					</Link>
					<Link href="/browse">
						<Button variant="outline" size="lg" className="px-8 h-11 text-sm font-medium">
							Browse Skills
						</Button>
					</Link>
				</motion.div>

				{/* Product illustration */}
				<motion.div
					initial={{ opacity: 0, y: 24 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.7, delay: 0.4 }}
					className="grid gap-4 md:grid-cols-12"
				>
					<motion.div
						initial={{ opacity: 0, x: -20 }}
						animate={{ opacity: 1, x: 0 }}
						transition={{ duration: 0.6, delay: 0.55 }}
						className="md:col-span-5"
					>
						<div className="rounded-2xl border border-border/80 bg-card/90 p-4 shadow-[0_18px_42px_-30px_rgba(30,22,12,0.42)] backdrop-blur-sm">
							<div className="mb-4 flex items-center justify-between border-b border-border/70 pb-3">
								<div>
									<p className="text-xs uppercase tracking-[0.15em] text-muted-foreground">Swap Request</p>
									<h3 className="text-sm font-semibold text-foreground">Portrait Photography &lt;-&gt; React Mentoring</h3>
								</div>
								<span className="rounded-full bg-primary/12 px-2.5 py-1 text-[10px] font-medium text-foreground">Matched</span>
							</div>

							<div className="space-y-3 text-sm">
								<div className="rounded-xl border border-border/70 bg-background px-3 py-2">
									<p className="text-[11px] text-muted-foreground">You teach</p>
									<p className="font-medium text-foreground">React performance and state architecture</p>
								</div>
								<div className="rounded-xl border border-border/70 bg-background px-3 py-2">
									<p className="text-[11px] text-muted-foreground">You learn</p>
									<p className="font-medium text-foreground">Portrait lighting + retouching workflows</p>
								</div>

								<div className="flex items-center gap-2 rounded-xl border border-border/70 bg-background px-3 py-2">
									<CalendarDays className="h-4 w-4 text-muted-foreground" />
									<p className="text-[13px] text-foreground">Thu, 7:30 PM - 45 min session</p>
								</div>
							</div>
						</div>
					</motion.div>

					<motion.div
						initial={{ opacity: 0, x: 20 }}
						animate={{ opacity: 1, x: 0 }}
						transition={{ duration: 0.6, delay: 0.65 }}
						className="md:col-span-7"
					>
						<div className="rounded-2xl border border-border/80 bg-background/95 p-4 shadow-[0_20px_40px_-28px_rgba(58,39,18,0.35)]">
							<div className="mb-4 flex items-center justify-between border-b border-border/70 pb-3">
								<div className="flex items-center gap-2 text-sm font-medium text-foreground">
									<MessageCircleMore className="h-4 w-4" />
									Session Chat
								</div>
								<div className="flex items-center gap-2 rounded-full border border-border px-2 py-1 text-[10px] text-muted-foreground">
									<Sparkles className="h-3 w-3" />
									E2EE Enabled
								</div>
							</div>

							<div className="space-y-2.5 text-sm">
								<div className="max-w-[85%] rounded-2xl rounded-bl-md bg-muted px-3 py-2 text-foreground">
									I can walk you through lighting ratios with my setup. Want a quick reference shot?
								</div>
								<div className="ml-auto max-w-[85%] rounded-2xl rounded-br-md bg-primary px-3 py-2 text-primary-foreground">
									Perfect. I will share a component profile template for our React performance session too.
								</div>
								<div className="max-w-[85%] rounded-2xl rounded-bl-md bg-muted px-3 py-2 text-foreground">
									Great, start video when ready.
								</div>
							</div>

							<div className="mt-4 flex items-center justify-between rounded-xl border border-border/70 bg-card px-3 py-2.5 text-xs">
								<span className="text-muted-foreground">Draft message: Uploading sample portraits now...</span>
								<span className="inline-flex items-center gap-1 text-foreground">
									<Video className="h-3.5 w-3.5" />
									Start Call
								</span>
							</div>
						</div>
					</motion.div>
				</motion.div>

				<p className="mt-6 text-xs text-muted-foreground md:text-sm">
					Set up a swap, align in chat, then jump straight into a live session.
				</p>
			</div>
		</section>
	);
};

export default HeroSection;
