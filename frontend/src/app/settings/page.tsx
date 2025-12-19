'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Camera, Save, Eye, EyeOff } from 'lucide-react';
import { getUserById } from '@/lib/dummy-data';

export default function SettingsPage() {
	const [activeTab, setActiveTab] = useState('profile');
	const [showPassword, setShowPassword] = useState(false);
	const [showConfirmPassword, setShowConfirmPassword] = useState(false);

	const currentUserId = '1'; // Sarah Johnson
	const user = getUserById(currentUserId);

	const [profileData, setProfileData] = useState({
		name: user?.name || '',
		email: user?.email || '',
		location: user?.location || '',
		bio: 'Passionate about teaching and learning new skills. Love connecting with people through knowledge exchange.',
		isPublic: user?.isPublic || true
	});

	const [passwordData, setPasswordData] = useState({
		currentPassword: '',
		newPassword: '',
		confirmPassword: ''
	});

	const [notificationSettings, setNotificationSettings] = useState({
		swapRequests: true,
		messages: true,
		ratings: true,
		emailNotifications: true,
		pushNotifications: true
	});

	const [privacySettings, setPrivacySettings] = useState({
		showEmail: false,
		showLocation: true,
		showSkills: true,
		allowMessages: true,
		allowSwapRequests: true
	});

	const handleProfileSave = () => {
		console.log('Saving profile:', profileData);
		// In a real app, this would update the user profile
	};

	const handlePasswordChange = () => {
		if (passwordData.newPassword !== passwordData.confirmPassword) {
			alert('New passwords do not match');
			return;
		}
		console.log('Changing password:', passwordData);
		// In a real app, this would change the password
	};

	const toggleNotification = (key: keyof typeof notificationSettings) => {
		setNotificationSettings((prev) => ({
			...prev,
			[key]: !prev[key]
		}));
	};

	const togglePrivacy = (key: keyof typeof privacySettings) => {
		setPrivacySettings((prev) => ({
			...prev,
			[key]: !prev[key]
		}));
	};

	if (!user) {
		return (
			<div className="min-h-screen bg-background flex items-center justify-center">
				<div className="text-center">
					<h1 className="text-2xl font-bold text-foreground mb-2">User not found</h1>
					<p className="text-muted-foreground">Unable to load user settings.</p>
				</div>
			</div>
		);
	}

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
				<div className="mb-8">
					<h1 className="text-3xl font-bold text-foreground mb-2">Settings</h1>
					<p className="text-muted-foreground">
						Manage your account settings and preferences
					</p>
				</div>

				<Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
					<TabsList className="grid w-full grid-cols-4">
						<TabsTrigger value="profile">Profile</TabsTrigger>
						<TabsTrigger value="security">Security</TabsTrigger>
						<TabsTrigger value="notifications">Notifications</TabsTrigger>
						<TabsTrigger value="privacy">Privacy</TabsTrigger>
					</TabsList>

					<TabsContent value="profile" className="space-y-6">
						<Card>
							<CardHeader>
								<CardTitle>Profile Information</CardTitle>
								<CardDescription>
									Update your personal information and profile settings
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-6">
								{/* Profile Picture */}
								<div className="flex items-center space-x-4">
									<Avatar
										src={user.photoUrl || undefined}
										alt={user.name}
										fallback={user.name
											.split(' ')
											.map((n) => n[0])
											.join('')}
										className="w-20 h-20"
									/>
									<div>
										<Button variant="outline" size="sm">
											<Camera className="w-4 h-4 mr-2" />
											Change Photo
										</Button>
										<p className="text-sm text-muted-foreground mt-1">
											JPG, PNG or GIF. Max size 2MB.
										</p>
									</div>
								</div>

								{/* Basic Information */}
								<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
									<div>
										<Label htmlFor="name">Full Name</Label>
										<Input
											id="name"
											value={profileData.name}
											onChange={(e) =>
												setProfileData((prev) => ({
													...prev,
													name: e.target.value
												}))
											}
										/>
									</div>
									<div>
										<Label htmlFor="email">Email</Label>
										<Input
											id="email"
											type="email"
											value={profileData.email}
											onChange={(e) =>
												setProfileData((prev) => ({
													...prev,
													email: e.target.value
												}))
											}
										/>
									</div>
									<div>
										<Label htmlFor="location">Location</Label>
										<Input
											id="location"
											value={profileData.location}
											onChange={(e) =>
												setProfileData((prev) => ({
													...prev,
													location: e.target.value
												}))
											}
											placeholder="City, State"
										/>
									</div>
									<div>
										<Label htmlFor="bio">Bio</Label>
										<Input
											id="bio"
											value={profileData.bio}
											onChange={(e) =>
												setProfileData((prev) => ({
													...prev,
													bio: e.target.value
												}))
											}
											placeholder="Tell us about yourself"
										/>
									</div>
								</div>

								{/* Skills */}
								<div>
									<Label>Skills Offered</Label>
									<div className="flex flex-wrap gap-2 mt-2">
										{user.skillsOffered.map((skill) => (
											<Badge key={skill.skillId} variant="secondary">
												{skill.name}
											</Badge>
										))}
									</div>
								</div>

								<div>
									<Label>Skills Wanted</Label>
									<div className="flex flex-wrap gap-2 mt-2">
										{user.skillsWanted.map((skill) => (
											<Badge key={skill.skillId} variant="outline">
												{skill.name}
											</Badge>
										))}
									</div>
								</div>

								<Button onClick={handleProfileSave}>
									<Save className="w-4 h-4 mr-2" />
									Save Changes
								</Button>
							</CardContent>
						</Card>
					</TabsContent>

					<TabsContent value="security" className="space-y-6">
						<Card>
							<CardHeader>
								<CardTitle>Change Password</CardTitle>
								<CardDescription>
									Update your password to keep your account secure
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4">
								<div>
									<Label htmlFor="currentPassword">Current Password</Label>
									<div className="relative">
										<Input
											id="currentPassword"
											type={showPassword ? 'text' : 'password'}
											value={passwordData.currentPassword}
											onChange={(e) =>
												setPasswordData((prev) => ({
													...prev,
													currentPassword: e.target.value
												}))
											}
										/>
										<Button
											type="button"
											variant="ghost"
											size="sm"
											className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
											onClick={() => setShowPassword(!showPassword)}
										>
											{showPassword ? (
												<EyeOff className="w-4 h-4" />
											) : (
												<Eye className="w-4 h-4" />
											)}
										</Button>
									</div>
								</div>
								<div>
									<Label htmlFor="newPassword">New Password</Label>
									<Input
										id="newPassword"
										type="password"
										value={passwordData.newPassword}
										onChange={(e) =>
											setPasswordData((prev) => ({
												...prev,
												newPassword: e.target.value
											}))
										}
									/>
								</div>
								<div>
									<Label htmlFor="confirmPassword">Confirm New Password</Label>
									<div className="relative">
										<Input
											id="confirmPassword"
											type={showConfirmPassword ? 'text' : 'password'}
											value={passwordData.confirmPassword}
											onChange={(e) =>
												setPasswordData((prev) => ({
													...prev,
													confirmPassword: e.target.value
												}))
											}
										/>
										<Button
											type="button"
											variant="ghost"
											size="sm"
											className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
											onClick={() =>
												setShowConfirmPassword(!showConfirmPassword)
											}
										>
											{showConfirmPassword ? (
												<EyeOff className="w-4 h-4" />
											) : (
												<Eye className="w-4 h-4" />
											)}
										</Button>
									</div>
								</div>
								<Button onClick={handlePasswordChange}>Change Password</Button>
							</CardContent>
						</Card>
					</TabsContent>

					<TabsContent value="notifications" className="space-y-6">
						<Card>
							<CardHeader>
								<CardTitle>Notification Preferences</CardTitle>
								<CardDescription>
									Choose what notifications you want to receive
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4">
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Swap Requests</h3>
										<p className="text-sm text-muted-foreground">
											When someone wants to swap skills with you
										</p>
									</div>
									<Button
										variant={
											notificationSettings.swapRequests
												? 'default'
												: 'outline'
										}
										size="sm"
										onClick={() => toggleNotification('swapRequests')}
									>
										{notificationSettings.swapRequests ? 'On' : 'Off'}
									</Button>
								</div>
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Messages</h3>
										<p className="text-sm text-muted-foreground">
											When you receive new messages
										</p>
									</div>
									<Button
										variant={
											notificationSettings.messages ? 'default' : 'outline'
										}
										size="sm"
										onClick={() => toggleNotification('messages')}
									>
										{notificationSettings.messages ? 'On' : 'Off'}
									</Button>
								</div>
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Ratings & Reviews</h3>
										<p className="text-sm text-muted-foreground">
											When someone rates your skills
										</p>
									</div>
									<Button
										variant={
											notificationSettings.ratings ? 'default' : 'outline'
										}
										size="sm"
										onClick={() => toggleNotification('ratings')}
									>
										{notificationSettings.ratings ? 'On' : 'Off'}
									</Button>
								</div>
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Email Notifications</h3>
										<p className="text-sm text-muted-foreground">
											Receive notifications via email
										</p>
									</div>
									<Button
										variant={
											notificationSettings.emailNotifications
												? 'default'
												: 'outline'
										}
										size="sm"
										onClick={() => toggleNotification('emailNotifications')}
									>
										{notificationSettings.emailNotifications ? 'On' : 'Off'}
									</Button>
								</div>
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Push Notifications</h3>
										<p className="text-sm text-muted-foreground">
											Receive notifications in your browser
										</p>
									</div>
									<Button
										variant={
											notificationSettings.pushNotifications
												? 'default'
												: 'outline'
										}
										size="sm"
										onClick={() => toggleNotification('pushNotifications')}
									>
										{notificationSettings.pushNotifications ? 'On' : 'Off'}
									</Button>
								</div>
							</CardContent>
						</Card>
					</TabsContent>

					<TabsContent value="privacy" className="space-y-6">
						<Card>
							<CardHeader>
								<CardTitle>Privacy Settings</CardTitle>
								<CardDescription>
									Control who can see your information
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4">
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Show Email</h3>
										<p className="text-sm text-muted-foreground">
											Allow others to see your email address
										</p>
									</div>
									<Button
										variant={privacySettings.showEmail ? 'default' : 'outline'}
										size="sm"
										onClick={() => togglePrivacy('showEmail')}
									>
										{privacySettings.showEmail ? 'Public' : 'Private'}
									</Button>
								</div>
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Show Location</h3>
										<p className="text-sm text-muted-foreground">
											Display your location on your profile
										</p>
									</div>
									<Button
										variant={
											privacySettings.showLocation ? 'default' : 'outline'
										}
										size="sm"
										onClick={() => togglePrivacy('showLocation')}
									>
										{privacySettings.showLocation ? 'Public' : 'Private'}
									</Button>
								</div>
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Show Skills</h3>
										<p className="text-sm text-muted-foreground">
											Display your skills on your profile
										</p>
									</div>
									<Button
										variant={privacySettings.showSkills ? 'default' : 'outline'}
										size="sm"
										onClick={() => togglePrivacy('showSkills')}
									>
										{privacySettings.showSkills ? 'Public' : 'Private'}
									</Button>
								</div>
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Allow Messages</h3>
										<p className="text-sm text-muted-foreground">
											Let others send you messages
										</p>
									</div>
									<Button
										variant={
											privacySettings.allowMessages ? 'default' : 'outline'
										}
										size="sm"
										onClick={() => togglePrivacy('allowMessages')}
									>
										{privacySettings.allowMessages ? 'Allowed' : 'Blocked'}
									</Button>
								</div>
								<div className="flex items-center justify-between">
									<div>
										<h3 className="font-medium">Allow Swap Requests</h3>
										<p className="text-sm text-muted-foreground">
											Let others send you swap requests
										</p>
									</div>
									<Button
										variant={
											privacySettings.allowSwapRequests
												? 'default'
												: 'outline'
										}
										size="sm"
										onClick={() => togglePrivacy('allowSwapRequests')}
									>
										{privacySettings.allowSwapRequests ? 'Allowed' : 'Blocked'}
									</Button>
								</div>
							</CardContent>
						</Card>
					</TabsContent>
				</Tabs>
			</div>
		</div>
	);
}
