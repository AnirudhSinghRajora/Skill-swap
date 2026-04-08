import { LoginForm } from '@/components/login-form';

export default function LoginPage() {
	return (
		<div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10 bg-background">
			<LoginForm />
		</div>
	);
}
