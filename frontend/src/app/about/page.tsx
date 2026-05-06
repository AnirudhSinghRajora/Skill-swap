import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Users, Star, Shield, Heart, ArrowRight, CheckCircle } from 'lucide-react';
import Link from 'next/link';

export default function AboutPage() {
	const features = [
		{
			icon: Users,
			title: 'Skill Exchange',
			description: 'Connect with people in your area to exchange skills and knowledge'
		},
		{
			icon: Star,
			title: 'Verified Users',
			description: 'All users are verified to ensure safe and quality exchanges'
		},
		{
			icon: Shield,
			title: 'Safe & Secure',
			description: 'Built-in safety features and community guidelines'
		},
		{
			icon: Heart,
			title: 'Community Driven',
			description: 'Join a community of learners and teachers'
		}
	];



	const howItWorks = [
		{
			step: '1',
			title: 'Create Your Profile',
			description: 'Sign up and list the skills you can offer and want to learn'
		},
		{
			step: '2',
			title: 'Find People',
			description: 'Browse profiles and find people with matching skills'
		},
		{
			step: '3',
			title: 'Make a Swap',
			description: 'Send a swap request and arrange your skill exchange'
		},
		{
			step: '4',
			title: 'Learn & Teach',
			description: 'Meet up and exchange your skills in person'
		}
	];

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				{/* Hero Section */}
				<div className="text-center mb-16">
					<h1 className="text-4xl font-bold text-foreground mb-4">About SkillSwap</h1>
					<p className="text-xl text-muted-foreground max-w-3xl mx-auto mb-8">
						SkillSwap is a platform that connects people who want to learn new skills
						with those who can teach them. We believe that everyone has something
						valuable to share and something new to learn.
					</p>
					<div className="flex flex-wrap justify-center gap-4">
					<Link href="/auth/signup">
						<Button size="lg">
							Get Started
							<ArrowRight className="w-4 h-4 ml-2" />
						</Button>
					</Link>
					<Link href="/browse">
						<Button variant="outline" size="lg">
							Browse People
						</Button>
					</Link>
				</div>
			</div>
				{/* Features */}
				<div className="mb-16">
					<h2 className="text-3xl font-bold text-foreground text-center mb-12">
						Why Choose SkillSwap?
					</h2>
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
						{features.map((feature) => {
							const Icon = feature.icon;
							return (
								<Card key={feature.title} className="text-center">
									<CardContent className="p-6">
										<div className="w-12 h-12 mx-auto mb-4 bg-primary/10 rounded-lg flex items-center justify-center">
											<Icon className="w-6 h-6 text-primary" />
										</div>
										<h3 className="font-semibold text-lg mb-2">
											{feature.title}
										</h3>
										<p className="text-muted-foreground text-sm">
											{feature.description}
										</p>
									</CardContent>
								</Card>
							);
						})}
					</div>
				</div>

				{/* How It Works */}
				<div className="mb-16">
					<h2 className="text-3xl font-bold text-foreground text-center mb-12">
						How It Works
					</h2>
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
						{howItWorks.map((step) => (
							<Card key={step.step}>
								<CardContent className="p-6 text-center">
									<div className="w-12 h-12 mx-auto mb-4 bg-primary rounded-full flex items-center justify-center">
										<span className="text-primary-foreground font-bold">
											{step.step}
										</span>
									</div>
									<h3 className="font-semibold text-lg mb-2">{step.title}</h3>
									<p className="text-muted-foreground text-sm">
										{step.description}
									</p>
								</CardContent>
							</Card>
						))}
					</div>
				</div>

				{/* Popular Skills */}
				<div className="mb-16">
					<h2 className="text-3xl font-bold text-foreground text-center mb-12">
						Popular Skills on SkillSwap
					</h2>
					<div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
						{[
							'JavaScript',
							'Cooking',
							'Photography',
							'Spanish',
							'Guitar',
							'Yoga',
							'Python',
							'Graphic Design',
							'Chess',
							'Woodworking',
							'French',
							'Piano'
						].map((skill) => (
							<div
								key={skill}
								className="text-center p-4 border border-border rounded-lg hover:border-primary/50 transition-colors"
							>
								<h3 className="font-medium text-sm">{skill}</h3>
							</div>
						))}
					</div>
				</div>

				{/* Safety & Guidelines */}
				<div className="mb-16">
					<h2 className="text-3xl font-bold text-foreground text-center mb-12">
						Safety & Community Guidelines
					</h2>
					<div className="grid grid-cols-1 md:grid-cols-2 gap-8">
						<Card>
							<CardHeader>
								<CardTitle>Safety First</CardTitle>
							</CardHeader>
							<CardContent className="space-y-3">
								<div className="flex items-start space-x-3">
									<CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
									<div>
										<h4 className="font-medium">Meet in Public Places</h4>
										<p className="text-sm text-muted-foreground">
											Always arrange to meet in public, well-lit locations
										</p>
									</div>
								</div>
								<div className="flex items-start space-x-3">
									<CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
									<div>
										<h4 className="font-medium">Verify Profiles</h4>
										<p className="text-sm text-muted-foreground">
											Check user ratings and reviews before meeting
										</p>
									</div>
								</div>
								<div className="flex items-start space-x-3">
									<CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
									<div>
										<h4 className="font-medium">Trust Your Instincts</h4>
										<p className="text-sm text-muted-foreground">
											If something feels wrong, don&apos;t proceed with the
											exchange
										</p>
									</div>
								</div>
							</CardContent>
						</Card>

						<Card>
							<CardHeader>
								<CardTitle>Community Guidelines</CardTitle>
							</CardHeader>
							<CardContent className="space-y-3">
								<div className="flex items-start space-x-3">
									<CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
									<div>
										<h4 className="font-medium">Be Respectful</h4>
										<p className="text-sm text-muted-foreground">
											Treat everyone with kindness and respect
										</p>
									</div>
								</div>
								<div className="flex items-start space-x-3">
									<CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
									<div>
										<h4 className="font-medium">Be Honest</h4>
										<p className="text-sm text-muted-foreground">
											Only offer skills you&apos;re confident teaching
										</p>
									</div>
								</div>
								<div className="flex items-start space-x-3">
									<CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
									<div>
										<h4 className="font-medium">Communicate Clearly</h4>
										<p className="text-sm text-muted-foreground">
											Set clear expectations for your skill exchanges
										</p>
									</div>
								</div>
							</CardContent>
						</Card>
					</div>
				</div>

				{/* CTA */}
				<Card className="text-center">
					<CardContent className="p-12">
						<h2 className="text-3xl font-bold text-foreground mb-4">
							Ready to Start Learning?
						</h2>
						<p className="text-muted-foreground mb-8 max-w-2xl mx-auto">
						Join our growing community of people exchanging skills and building
						meaningful connections.
					</p>
					<div className="flex flex-wrap justify-center gap-4">
						<Link href="/auth/signup">
								<Button size="lg">
									Create Account
									<ArrowRight className="w-4 h-4 ml-2" />
								</Button>
							</Link>
							<Link href="/browse">
								<Button variant="outline" size="lg">
									Explore Skills
								</Button>
							</Link>
						</div>
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
