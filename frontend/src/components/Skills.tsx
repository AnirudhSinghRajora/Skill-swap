'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { useInView } from 'framer-motion';
import { useRef } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
	Code,
	Palette,
	BarChart,
	Megaphone,
	Camera,
	Music,
	Languages,
	Wrench,
	ArrowRight
} from 'lucide-react';

const SkillsMarketplace = () => {
	const [selectedCategory, setSelectedCategory] = useState('All');
	const ref = useRef(null);
	const inView = useInView(ref, { once: true, amount: 0.1 });

	const skillCategories = [
		{ icon: Code, name: 'Programming', count: 156 },
		{ icon: Palette, name: 'Design', count: 124 },
		{ icon: BarChart, name: 'Analytics', count: 89 },
		{ icon: Megaphone, name: 'Marketing', count: 112 },
		{ icon: Camera, name: 'Photography', count: 67 },
		{ icon: Music, name: 'Music', count: 45 },
		{ icon: Languages, name: 'Languages', count: 78 },
		{ icon: Wrench, name: 'Engineering', count: 134 }
	];

	const categories = ['All', ...skillCategories.map((cat) => cat.name)];

	const filteredCategories =
		selectedCategory === 'All'
			? skillCategories
			: skillCategories.filter((cat) => cat.name === selectedCategory);

	const trendingSkills = [
		'React.js', 'UI/UX Design', 'Python', 'Digital Marketing',
		'Photography', 'Data Science', 'Spanish', 'Video Editing',
		'Machine Learning', 'Graphic Design'
	];

	return (
		<section className="py-24 md:py-32 border-t border-border bg-muted/30">
			<div ref={ref} className="max-w-6xl mx-auto px-4">
				{/* Header */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6 }}
					className="max-w-2xl mb-12"
				>
					<p className="text-sm font-medium text-primary mb-3 tracking-wide uppercase">Marketplace</p>
					<h2 className="text-display-lg text-foreground mb-4">
						Explore skill categories
					</h2>
					<p className="text-lg text-muted-foreground leading-relaxed">
						From programming to photography — find the right exchange partner.
					</p>
				</motion.div>

				{/* Category Filter */}
				<motion.div
					initial={{ opacity: 0, y: 16 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.5, delay: 0.1 }}
					className="flex flex-wrap gap-2 mb-12"
				>
					{categories.map((category) => (
						<button
							key={category}
							onClick={() => setSelectedCategory(category)}
							className={`text-sm px-4 py-1.5 rounded-full border transition-colors ${
								selectedCategory === category
									? 'bg-foreground text-background border-foreground'
									: 'bg-transparent text-muted-foreground border-border hover:border-foreground/30'
							}`}
						>
							{category}
						</button>
					))}
				</motion.div>

				{/* Skills Grid */}
				<div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-20">
					{filteredCategories.map((skill, index) => (
						<motion.div
							key={skill.name}
							initial={{ opacity: 0, y: 16 }}
							animate={inView ? { opacity: 1, y: 0 } : {}}
							transition={{ duration: 0.4, delay: 0.05 * index }}
							className="group p-6 rounded-xl border border-border bg-card hover:border-foreground/20 transition-colors cursor-pointer"
						>
							<skill.icon className="h-5 w-5 text-muted-foreground mb-4 group-hover:text-foreground transition-colors" strokeWidth={1.5} />
							<h3 className="font-medium text-foreground mb-1">{skill.name}</h3>
							<p className="text-sm text-muted-foreground">{skill.count} professionals</p>
						</motion.div>
					))}
				</div>

				{/* Trending */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, delay: 0.3 }}
					className="text-center"
				>
					<p className="text-sm font-medium text-muted-foreground mb-6">Trending this week</p>
					<div className="flex flex-wrap justify-center gap-2 mb-10">
						{trendingSkills.map((skill) => (
							<Badge
								key={skill}
								variant="outline"
								className="px-4 py-1.5 text-sm cursor-pointer hover:bg-accent transition-colors rounded-full"
							>
								{skill}
							</Badge>
						))}
					</div>

					<Button variant="outline" className="group">
						Browse all skills
						<ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
					</Button>
				</motion.div>
			</div>
		</section>
	);
};

export default SkillsMarketplace;
