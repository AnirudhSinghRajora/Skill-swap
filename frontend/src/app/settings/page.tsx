'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, Check, Clock, Trash2, Plus } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api, { AvailabilitySlot } from '@/lib/api';
import { DAYS, getDayLabels, formatTime } from '@/types/availability';
import { toast } from 'sonner';

export default function SettingsPage() {
	const { user, isLoading: authLoading } = useAuth(true);
	const queryClient = useQueryClient();
	const [saveMessage, setSaveMessage] = useState('');

	// Availability
	const { data: slots = [], isLoading: slotsLoading } = useQuery({
		queryKey: ['availability'],
		queryFn: () => api.availability.list(),
		enabled: !!user,
	});

	const [showSlotForm, setShowSlotForm] = useState(false);
	const [slotLabel, setSlotLabel] = useState('');
	const [slotDays, setSlotDays] = useState(0);
	const [slotStart, setSlotStart] = useState('09:00');
	const [slotEnd, setSlotEnd] = useState('17:00');

	const createSlotMutation = useMutation({
		mutationFn: (data: { label: string; day_bitmask: number; start_time: string; end_time: string }) =>
			api.availability.create(data),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['availability'] });
			setShowSlotForm(false);
			setSlotLabel('');
			setSlotDays(0);
			setSlotStart('09:00');
			setSlotEnd('17:00');
			setSaveMessage('Availability slot added!');
			setTimeout(() => setSaveMessage(''), 3000);
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to add availability slot'),
	});

	const deleteSlotMutation = useMutation({
		mutationFn: (id: string) => api.availability.delete(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['availability'] });
			setSaveMessage('Availability slot removed.');
			setTimeout(() => setSaveMessage(''), 3000);
		},
		onError: (err: Error) => toast.error(err.message || 'Failed to remove availability slot'),
	});

	const toggleDay = (bit: number) => {
		setSlotDays((prev) => prev ^ bit);
	};

	const handleCreateSlot = () => {
		if (!slotLabel.trim() || slotDays === 0) return;
		createSlotMutation.mutate({
			label: slotLabel.trim(),
			day_bitmask: slotDays,
			start_time: slotStart + ':00',
			end_time: slotEnd + ':00',
		});
	};

	if (authLoading || slotsLoading) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
			</div>
		);
	}

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				<div className="mb-10 animate-fade-in-up">
					<h1 className="text-display-md text-foreground mb-2">Settings</h1>
					<p className="text-muted-foreground text-lg">
						Manage your availability schedule
					</p>
				</div>

				{saveMessage && (
					<div className="mb-6 p-3 bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded-lg flex items-center gap-2 text-green-700 dark:text-green-300">
						<Check className="w-4 h-4" />
						{saveMessage}
					</div>
				)}

				<Card>
					<CardHeader>
						<div className="flex items-center justify-between">
							<div>
								<CardTitle>Availability Slots</CardTitle>
								<CardDescription>
									Set your available times so others can find matching schedules
								</CardDescription>
							</div>
							{!showSlotForm && (
								<Button size="sm" onClick={() => setShowSlotForm(true)}>
									<Plus className="w-4 h-4 mr-2" />
									Add Slot
								</Button>
							)}
						</div>
					</CardHeader>
					<CardContent className="space-y-4">
						{showSlotForm && (
							<div className="p-4 border border-border rounded-lg space-y-4">
								<div>
									<Label htmlFor="slot-label">Label</Label>
									<Input
										id="slot-label"
										value={slotLabel}
										onChange={(e) => setSlotLabel(e.target.value)}
										placeholder="e.g. Weekday Evenings"
									/>
								</div>
								<div>
									<Label>Days</Label>
									<div className="flex flex-wrap gap-2 mt-1">
										{DAYS.map((day) => (
											<Button
												key={day.bit}
												type="button"
												size="sm"
												variant={slotDays & day.bit ? 'default' : 'outline'}
												onClick={() => toggleDay(day.bit)}
											>
												{day.label}
											</Button>
										))}
									</div>
								</div>
								<div className="grid grid-cols-2 gap-4">
									<div>
										<Label htmlFor="slot-start">Start Time</Label>
										<Input
											id="slot-start"
											type="time"
											value={slotStart}
											onChange={(e) => setSlotStart(e.target.value)}
										/>
									</div>
									<div>
										<Label htmlFor="slot-end">End Time</Label>
										<Input
											id="slot-end"
											type="time"
											value={slotEnd}
											onChange={(e) => setSlotEnd(e.target.value)}
										/>
									</div>
								</div>
								<div className="flex gap-2">
									<Button
										onClick={handleCreateSlot}
										disabled={createSlotMutation.isPending || !slotLabel.trim() || slotDays === 0}
									>
										{createSlotMutation.isPending ? (
											<Loader2 className="w-4 h-4 mr-2 animate-spin" />
										) : (
											<Plus className="w-4 h-4 mr-2" />
										)}
										Save Slot
									</Button>
									<Button variant="outline" onClick={() => setShowSlotForm(false)}>
										Cancel
									</Button>
								</div>
							</div>
						)}

						{slots.length === 0 && !showSlotForm ? (
							<p className="text-center text-muted-foreground py-8">
								No availability slots set. Add one so others can find times that work for both of you.
							</p>
						) : (
							slots.map((slot: AvailabilitySlot) => (
								<div
									key={slot.slot_id}
									className="flex items-center justify-between p-4 border border-border rounded-lg"
								>
									<div className="space-y-1">
										<div className="flex items-center gap-2">
											<Clock className="w-4 h-4 text-muted-foreground" />
											<p className="font-medium">{slot.label}</p>
										</div>
										<div className="flex flex-wrap gap-1">
											{getDayLabels(slot.day_bitmask).map((day) => (
												<span
													key={day}
													className="text-xs px-2 py-0.5 bg-primary/10 text-primary rounded-full"
												>
													{day}
												</span>
											))}
										</div>
										<p className="text-sm text-muted-foreground">
											{formatTime(slot.start_time)} – {formatTime(slot.end_time)}
										</p>
									</div>
									<Button
										variant="ghost"
										size="sm"
										onClick={() => deleteSlotMutation.mutate(slot.slot_id)}
										disabled={deleteSlotMutation.isPending}
									>
										<Trash2 className="w-4 h-4 text-destructive" />
									</Button>
								</div>
							))
						)}
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
