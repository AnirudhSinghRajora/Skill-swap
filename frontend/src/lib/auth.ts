import { NextAuthOptions, User as NextAuthUser } from 'next-auth';
import GitHubProvider from 'next-auth/providers/github';
import CredentialsProvider from 'next-auth/providers/credentials';

type BackendUserResponse = {
	id: string;
	name?: string;
	email: string;
	role?: string;
};

declare module 'next-auth' {
	interface Session {
		user: {
			id: string;
			name?: string;
			email?: string;
			role?: string;
		};
	}

	interface User {
		id: string;
		name?: string;
		email?: string;
		role?: string;
	}
}

declare module 'next-auth/jwt' {
	interface JWT {
		id?: string;
		name?: string;
		email?: string;
		role?: string;
	}
}

export const authOptions: NextAuthOptions = {
	providers: [
		GitHubProvider({
			clientId: process.env.GITHUB_ID!,
			clientSecret: process.env.GITHUB_SECRET!
		}),

		CredentialsProvider({
			name: 'Credentials',
			credentials: {
				email: { label: 'Email', type: 'text' },
				password: { label: 'Password', type: 'password' }
			},
			async authorize(credentials): Promise<NextAuthUser | null> {
				const { email, password } = credentials ?? {};

				if (!email || !password) {
					return null;
				}

				const res = await fetch(`${process.env.GO_BACKEND_URL}/api/auth/login`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({ email, password })
				});

				if (!res.ok) {
					return null;
				}

				const user: BackendUserResponse = await res.json();

				if (!user?.id || !user?.email) {
					return null;
				}

				return {
					id: user.id,
					name: user.name ?? undefined,
					email: user.email,
					role: user.role ?? 'user'
				};
			}
		})
	],

	session: {
		strategy: 'jwt',
		maxAge: 30 * 24 * 60 * 60
	},

	callbacks: {
		async jwt({ token, user }) {
			if (user) {
				token.id = user.id;
				token.name = user.name;
				token.email = user.email;
				token.role = user.role;
			}
			return token;
		},

		async session({ session, token }) {
			if (session.user) {
				session.user.id = token.id!;
				session.user.name = token.name;
				session.user.email = token.email;
				session.user.role = token.role;
			}
			return session;
		}
	},

	pages: {
		signIn: '/auth/signin',
		error: '/auth/error'
	},

	secret: process.env.NEXTAUTH_SECRET!
};
