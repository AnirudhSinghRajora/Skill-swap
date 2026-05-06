export type NotificationType =
	| 'swap_request'
	| 'swap_accepted'
	| 'swap_rejected'
	| 'swap_completed'
	| 'new_rating'
	| 'new_message'
	| 'skill_matched'
	| 'system_alert'
	| 'admin_notice';

export interface NotificationItem {
	notification_id: string;
	type: NotificationType;
	title: string;
	message: string;
	is_read: boolean;
	related_id: string | null;
	created_at: string;
}

export interface NotificationListResponse {
	notifications: NotificationItem[];
	pagination: {
		page: number;
		limit: number;
		total: number;
	};
}

export interface NotificationStats {
	total_notifications: number;
	unread_count: number;
	read_count: number;
}
