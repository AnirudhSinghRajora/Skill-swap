import HeroSection from '@/components/Hero';
import FeaturesSection from '@/components/Features';
import SkillsMarketplace from '@/components/Skills';
import HowItWorksSection from '@/components/HowItWorks';
import CallToActionSection from '@/components/CTA';

export default function Home() {
	return (
		<main className="min-h-screen">
			<HeroSection />
			<FeaturesSection />
			<SkillsMarketplace />
			<HowItWorksSection />
			<CallToActionSection />
		</main>
	);
}
