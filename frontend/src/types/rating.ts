export interface RatingType {
	rating_id: string;
	swap_id: string;
	rater_id: string;
	ratee_id: string;
	score: number;
	comment: string | null;
	created_at: string;
	rater?: { user_id: string; name: string; has_photo: boolean };
	ratee?: { user_id: string; name: string; has_photo: boolean };
}

export interface UserRatingStats {
	average_rating: number;
	total_ratings: number;
}
