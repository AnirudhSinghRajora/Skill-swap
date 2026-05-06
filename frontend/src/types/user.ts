import { SkillType } from './skill';

export interface UserType {
	user_id: string;
	name: string;
	email: string;
	location: string | null;
	has_photo: boolean;
	is_public: boolean;
	public_key?: string;
	has_key_backup?: boolean;
	skills_offered: SkillType[];
	skills_wanted: SkillType[];
	created_at: string;
}

export interface SearchUsersResponse {
	users: UserType[];
	total: number;
	page: number;
	limit: number;
	total_pages: number;
}
