import { UserType } from '@/types/user';
import { SkillType } from '@/types/skill';
import { SwapRequest } from '@/types/swapRequest';
import { RatingType } from '@/types/rating';
import { Notification } from '@/types/notification';
import { MessageType } from '@/types/message';

// Sample skills
export const dummySkills: SkillType[] = [
	{
		skillId: '1',
		name: 'JavaScript',
		description: 'Modern JavaScript programming and web development'
	},
	{
		skillId: '2',
		name: 'Python',
		description: 'Python programming for data science and automation'
	},
	{
		skillId: '3',
		name: 'Graphic Design',
		description: 'Adobe Creative Suite and digital design'
	},
	{
		skillId: '4',
		name: 'Cooking',
		description: 'International cuisine and culinary techniques'
	},
	{
		skillId: '5',
		name: 'Photography',
		description: 'Digital and film photography techniques'
	},
	{
		skillId: '6',
		name: 'Guitar',
		description: 'Acoustic and electric guitar lessons'
	},
	{
		skillId: '7',
		name: 'Spanish',
		description: 'Spanish language conversation and grammar'
	},
	{
		skillId: '8',
		name: 'Yoga',
		description: 'Vinyasa and Hatha yoga instruction'
	},
	{
		skillId: '9',
		name: 'Woodworking',
		description: 'Furniture making and wood crafting'
	},
	{
		skillId: '10',
		name: 'Chess',
		description: 'Chess strategy and advanced techniques'
	}
];

// Sample users
export const dummyUsers: UserType[] = [
	{
		userId: '1',
		name: 'Sarah Johnson',
		email: 'sarah.johnson@email.com',
		role: 'user',
		rating: 4.8,
		location: 'San Francisco, CA',
		photoUrl:
			'https://images.unsplash.com/photo-1494790108755-2616b612b786?w=150&h=150&fit=crop&crop=face',
		isPublic: true,
		createdAt: '2024-01-15T10:00:00Z',
		updatedAt: '2024-01-15T10:00:00Z',
		deletedAt: null,
		skillsOffered: [dummySkills[0], dummySkills[1]],
		skillsWanted: [dummySkills[3], dummySkills[6]]
	},
	{
		userId: '2',
		name: 'Michael Chen',
		email: 'michael.chen@email.com',
		role: 'user',
		rating: 4.6,
		location: 'New York, NY',
		photoUrl:
			'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150&h=150&fit=crop&crop=face',
		isPublic: true,
		createdAt: '2024-01-10T14:30:00Z',
		updatedAt: '2024-01-10T14:30:00Z',
		deletedAt: null,
		skillsOffered: [dummySkills[2], dummySkills[4]],
		skillsWanted: [dummySkills[0], dummySkills[7]]
	},
	{
		userId: '3',
		name: 'Emma Rodriguez',
		email: 'emma.rodriguez@email.com',
		role: 'user',
		rating: 4.9,
		location: 'Austin, TX',
		photoUrl:
			'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=150&h=150&fit=crop&crop=face',
		isPublic: true,
		createdAt: '2024-01-05T09:15:00Z',
		updatedAt: '2024-01-05T09:15:00Z',
		deletedAt: null,
		skillsOffered: [dummySkills[3], dummySkills[7]],
		skillsWanted: [dummySkills[1], dummySkills[5]]
	},
	{
		userId: '4',
		name: 'David Kim',
		email: 'david.kim@email.com',
		role: 'user',
		rating: 4.7,
		location: 'Seattle, WA',
		photoUrl:
			'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=150&h=150&fit=crop&crop=face',
		isPublic: true,
		createdAt: '2024-01-12T16:45:00Z',
		updatedAt: '2024-01-12T16:45:00Z',
		deletedAt: null,
		skillsOffered: [dummySkills[5], dummySkills[8]],
		skillsWanted: [dummySkills[2], dummySkills[4]]
	},
	{
		userId: '5',
		name: 'Lisa Thompson',
		email: 'lisa.thompson@email.com',
		role: 'user',
		rating: 4.5,
		location: 'Portland, OR',
		photoUrl:
			'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=150&h=150&fit=crop&crop=face',
		isPublic: true,
		createdAt: '2024-01-08T11:20:00Z',
		updatedAt: '2024-01-08T11:20:00Z',
		deletedAt: null,
		skillsOffered: [dummySkills[6], dummySkills[9]],
		skillsWanted: [dummySkills[0], dummySkills[3]]
	},
	{
		userId: '6',
		name: 'Alex Morgan',
		email: 'alex.morgan@email.com',
		role: 'user',
		rating: 4.4,
		location: 'Denver, CO',
		photoUrl:
			'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=150&h=150&fit=crop&crop=face',
		isPublic: true,
		createdAt: '2024-01-20T13:10:00Z',
		updatedAt: '2024-01-20T13:10:00Z',
		deletedAt: null,
		skillsOffered: [dummySkills[7], dummySkills[9]],
		skillsWanted: [dummySkills[1], dummySkills[8]]
	}
];

// Sample swap requests
export const dummySwapRequests: SwapRequest[] = [
	{
		swapId: '1',
		requesterId: '1',
		responderId: '2',
		offeredSkill: dummySkills[0],
		wantedSkill: dummySkills[2],
		status: 'pending',
		createdAt: '2024-01-25T10:00:00Z',
		updatedAt: '2024-01-25T10:00:00Z',
		deletedAt: null
	},
	{
		swapId: '2',
		requesterId: '3',
		responderId: '4',
		offeredSkill: dummySkills[3],
		wantedSkill: dummySkills[5],
		status: 'accepted',
		createdAt: '2024-01-23T14:30:00Z',
		updatedAt: '2024-01-24T09:15:00Z',
		deletedAt: null
	},
	{
		swapId: '3',
		requesterId: '5',
		responderId: '6',
		offeredSkill: dummySkills[6],
		wantedSkill: dummySkills[0],
		status: 'rejected',
		createdAt: '2024-01-22T16:45:00Z',
		updatedAt: '2024-01-23T11:20:00Z',
		deletedAt: null
	}
];

// Sample ratings
export const dummyRatings: RatingType[] = [
	{
		ratingId: '1',
		userId: '1',
		ratedById: '2',
		score: 5,
		comment: 'Excellent JavaScript teacher! Very patient and knowledgeable.',
		createdAt: '2024-01-20T10:00:00Z',
		updatedAt: '2024-01-20T10:00:00Z'
	},
	{
		ratingId: '2',
		userId: '2',
		ratedById: '1',
		score: 4,
		comment: 'Great graphic design skills. Helped me improve my portfolio.',
		createdAt: '2024-01-19T14:30:00Z',
		updatedAt: '2024-01-19T14:30:00Z'
	},
	{
		ratingId: '3',
		userId: '3',
		ratedById: '4',
		score: 5,
		comment: 'Amazing cooking instructor! Learned so much about Italian cuisine.',
		createdAt: '2024-01-18T16:45:00Z',
		updatedAt: '2024-01-18T16:45:00Z'
	}
];

// Sample notifications
export const dummyNotifications: Notification[] = [
	{
		notificationId: '1',
		type: 'swapRequest',
		content: 'Michael Chen wants to swap Graphic Design for JavaScript',
		isRead: false,
		createdAt: '2024-01-25T10:00:00Z',
		updatedAt: '2024-01-25T10:00:00Z'
	},
	{
		notificationId: '2',
		type: 'swapAccepted',
		content: 'Emma Rodriguez accepted your cooking for guitar swap',
		isRead: true,
		createdAt: '2024-01-24T09:15:00Z',
		updatedAt: '2024-01-24T09:15:00Z'
	},
	{
		notificationId: '3',
		type: 'messageReceived',
		content: 'New message from David Kim about your woodworking skills',
		isRead: false,
		createdAt: '2024-01-25T08:30:00Z',
		updatedAt: '2024-01-25T08:30:00Z'
	}
];

// Sample messages
export const dummyMessages: MessageType[] = [
	{
		messageId: '1',
		senderId: '1',
		receiverId: '2',
		text: 'Hi! I saw you offer graphic design. Would you be interested in swapping for JavaScript lessons?',
		image: null,
		createdAt: '2024-01-25T10:00:00Z',
		updatedAt: '2024-01-25T10:00:00Z'
	},
	{
		messageId: '2',
		senderId: '2',
		receiverId: '1',
		text: "Absolutely! I've been wanting to learn JavaScript. When would you be available?",
		image: null,
		createdAt: '2024-01-25T10:15:00Z',
		updatedAt: '2024-01-25T10:15:00Z'
	},
	{
		messageId: '3',
		senderId: '1',
		receiverId: '2',
		text: 'Great! How about this weekend? I can show you some portfolio examples too.',
		image: null,
		createdAt: '2024-01-25T10:30:00Z',
		updatedAt: '2024-01-25T10:30:00Z'
	}
];

// Helper functions
export const getUserById = (userId: string): UserType | undefined => {
	return dummyUsers.find((user) => user.userId === userId);
};

export const getSkillById = (skillId: string): SkillType | undefined => {
	return dummySkills.find((skill) => skill.skillId === skillId);
};

export const getSwapRequestsByUserId = (userId: string): SwapRequest[] => {
	return dummySwapRequests.filter(
		(request) => request.requesterId === userId || request.responderId === userId
	);
};

export const getRatingsByUserId = (userId: string): RatingType[] => {
	return dummyRatings.filter((rating) => rating.userId === userId);
};

export const getNotificationsByUserId = (): Notification[] => {
	return dummyNotifications;
};

export const getMessagesByUserId = (userId: string): MessageType[] => {
	return dummyMessages.filter(
		(message) => message.senderId === userId || message.receiverId === userId
	);
};
