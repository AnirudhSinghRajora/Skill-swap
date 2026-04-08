'use client';

import { useRef } from 'react';
import { motion, useInView } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { ArrowRight } from 'lucide-react';
import Link from 'next/link';

const CallToActionSection = () => {
	const ref = useRef(null);
	const inView = useInView(ref, { once: true, amount: 0.2 });

	return (
		<section className="border-t border-border">
			<div ref={ref} className="max-w-6xl mx-auto px-4 py-24 md:py-32">
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6 }}
					className="max-w-2xl mx-auto text-center"
				>
					<h2 className="text-display-lg text-foreground mb-4">
						Start exchanging skills today
					</h2>
					<p className="text-lg text-muted-foreground leading-relaxed mb-8">
						Join a community of professionals who teach what they know and learn what they need. Free to join, no credit card required.
					</p>

					<div className="flex flex-col sm:flex-row gap-3 justify-center mb-12">
						<Button size="lg" asChild>
							<Link href="/auth/signup">
								Create free account
								<ArrowRight className="ml-2 h-4 w-4" />
							</Link>
						</Button>
						<Button size="lg" variant="outline" asChild>
							<Link href="/auth/signin">Sign in</Link>
						</Button>
					</div>

					<div className="flex items-center justify-center gap-8 text-sm text-muted-foreground">
						<span>Free forever</span>
						<span className="w-1 h-1 rounded-full bg-border" />
						<span>No credit card</span>
						<span className="w-1 h-1 rounded-full bg-border" />
						<span>Cancel anytime</span>
					</div>
				</motion.div>
			</div>
		</section>
	);
};

export default CallToActionSection;
