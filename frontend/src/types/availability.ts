export interface AvailabilitySlotType {
	slot_id: string;
	user_id: string;
	label: string;
	day_bitmask: number;
	start_time: string;
	end_time: string;
	created_at: string;
}

// Day bitmask constants
export const DAYS = [
	{ bit: 1, label: 'Sun' },
	{ bit: 2, label: 'Mon' },
	{ bit: 4, label: 'Tue' },
	{ bit: 8, label: 'Wed' },
	{ bit: 16, label: 'Thu' },
	{ bit: 32, label: 'Fri' },
	{ bit: 64, label: 'Sat' },
] as const;

export function getDayLabels(bitmask: number): string[] {
	return DAYS.filter((d) => bitmask & d.bit).map((d) => d.label);
}

export function formatTime(timeStr: string): string {
	// Backend sends time as "0000-01-01T15:00:00Z" or "15:00:00"
	const match = timeStr.match(/(\d{2}):(\d{2})/);
	if (!match) return timeStr;
	const hours = parseInt(match[1], 10);
	const minutes = match[2];
	const ampm = hours >= 12 ? 'PM' : 'AM';
	const h = hours % 12 || 12;
	return `${h}:${minutes} ${ampm}`;
}
