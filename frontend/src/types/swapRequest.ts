import { SkillType } from './skill';
import { UserType } from './user';

export type SwapStatusType = 'pending' | 'accepted' | 'rejected' | 'cancelled' | 'completed';

export interface SwapRequestType {
	swap_id: string;
	requester_id: string;
	responder_id: string;
	offered_skill_id: string;
	wanted_skill_id: string;
	status: SwapStatusType;
	requester_completed: boolean;
	responder_completed: boolean;
	created_at: string;
	updated_at: string;
	requester: Pick<UserType, 'user_id' | 'name' | 'email' | 'location' | 'has_photo'>;
	responder: Pick<UserType, 'user_id' | 'name' | 'email' | 'location' | 'has_photo'>;
	offered_skill: SkillType;
	wanted_skill: SkillType;
}

export interface SwapListResponse {
	sent: SwapRequestType[];
	received: SwapRequestType[];
}
