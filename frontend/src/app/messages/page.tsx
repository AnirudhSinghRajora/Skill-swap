'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
// import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar } from '@/components/ui/avatar';
// import { Badge } from '@/components/ui/badge'
import {
	Search,
	Send,
	MoreVertical,
	Phone,
	Video,
	Image as ImageIcon,
	Paperclip
} from 'lucide-react';
import { dummyUsers, dummyMessages, getUserById } from '@/lib/dummy-data';

export default function MessagesPage() {
	const [selectedUser, setSelectedUser] = useState<string | null>('2'); // Michael Chen
	const [newMessage, setNewMessage] = useState('');
	const currentUserId = '1'; // Sarah Johnson

	// Get conversations (users who have exchanged messages with current user)
	const conversations = dummyUsers.filter(
		(user) =>
			user.userId !== currentUserId &&
			dummyMessages.some(
				(msg) =>
					(msg.senderId === currentUserId && msg.receiverId === user.userId) ||
					(msg.senderId === user.userId && msg.receiverId === currentUserId)
			)
	);

	// Get messages for selected conversation
	const conversationMessages = selectedUser
		? dummyMessages
				.filter(
					(msg) =>
						(msg.senderId === currentUserId && msg.receiverId === selectedUser) ||
						(msg.senderId === selectedUser && msg.receiverId === currentUserId)
				)
				.sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())
		: [];

	const selectedUserData = selectedUser ? getUserById(selectedUser) : null;

	const handleSendMessage = () => {
		if (newMessage.trim() && selectedUser) {
			// In a real app, this would send the message to the backend
			console.log('Sending message:', newMessage, 'to:', selectedUser);
			setNewMessage('');
		}
	};

	const handleKeyPress = (e: React.KeyboardEvent) => {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSendMessage();
		}
	};

	return (
		<div className="min-h-screen bg-background">
			<div className="max-w-6xl mx-auto h-screen flex">
				{/* Conversations List */}
				<div className="w-80 border-r border-border bg-card">
					<div className="p-4 border-b border-border">
						<h1 className="text-xl font-bold text-foreground mb-4">Messages</h1>
						<div className="relative">
							<Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
							<Input placeholder="Search conversations..." className="pl-10" />
						</div>
					</div>

					<div className="overflow-y-auto h-[calc(100vh-120px)]">
						{conversations.map((user) => {
							const lastMessage = dummyMessages
								.filter(
									(msg) =>
										(msg.senderId === currentUserId &&
											msg.receiverId === user.userId) ||
										(msg.senderId === user.userId &&
											msg.receiverId === currentUserId)
								)
								.sort(
									(a, b) =>
										new Date(b.createdAt).getTime() -
										new Date(a.createdAt).getTime()
								)[0];

							return (
								<div
									key={user.userId}
									className={`p-4 border-b border-border cursor-pointer hover:bg-muted/50 transition-colors ${
										selectedUser === user.userId ? 'bg-muted' : ''
									}`}
									onClick={() => setSelectedUser(user.userId)}
								>
									<div className="flex items-center space-x-3">
										<Avatar
											src={user.photoUrl || undefined}
											alt={user.name}
											fallback={user.name
												.split(' ')
												.map((n) => n[0])
												.join('')}
											className="w-12 h-12"
										/>
										<div className="flex-1 min-w-0">
											<div className="flex items-center justify-between">
												<h3 className="font-medium text-foreground truncate">
													{user.name}
												</h3>
												{lastMessage && (
													<span className="text-xs text-muted-foreground">
														{new Date(
															lastMessage.createdAt
														).toLocaleDateString()}
													</span>
												)}
											</div>
											{lastMessage && (
												<p className="text-sm text-muted-foreground truncate">
													{lastMessage.senderId === currentUserId
														? 'You: '
														: ''}
													{lastMessage.text}
												</p>
											)}
										</div>
									</div>
								</div>
							);
						})}
					</div>
				</div>

				{/* Chat Area */}
				<div className="flex-1 flex flex-col">
					{selectedUserData ? (
						<>
							{/* Chat Header */}
							<div className="p-4 border-b border-border bg-card flex items-center justify-between">
								<div className="flex items-center space-x-3">
									<Avatar
										src={selectedUserData.photoUrl || undefined}
										alt={selectedUserData.name}
										fallback={selectedUserData.name
											.split(' ')
											.map((n) => n[0])
											.join('')}
										className="w-10 h-10"
									/>
									<div>
										<h2 className="font-semibold text-foreground">
											{selectedUserData.name}
										</h2>
										<p className="text-sm text-muted-foreground">
											{selectedUserData.location}
										</p>
									</div>
								</div>
								<div className="flex items-center space-x-2">
									<Button variant="ghost" size="sm">
										<Phone className="w-4 h-4" />
									</Button>
									<Button variant="ghost" size="sm">
										<Video className="w-4 h-4" />
									</Button>
									<Button variant="ghost" size="sm">
										<MoreVertical className="w-4 h-4" />
									</Button>
								</div>
							</div>

							{/* Messages */}
							<div className="flex-1 overflow-y-auto p-4 space-y-4">
								{conversationMessages.map((message) => {
									const isOwnMessage = message.senderId === currentUserId;
									const sender = getUserById(message.senderId);

									return (
										<div
											key={message.messageId}
											className={`flex ${isOwnMessage ? 'justify-end' : 'justify-start'}`}
										>
											<div
												className={`max-w-xs lg:max-w-md ${isOwnMessage ? 'order-2' : 'order-1'}`}
											>
												{!isOwnMessage && (
													<div className="flex items-center space-x-2 mb-1">
														<Avatar
															src={sender?.photoUrl || undefined}
															alt={sender?.name || 'User'}
															fallback={
																sender?.name
																	?.split(' ')
																	.map((n) => n[0])
																	.join('') || 'U'
															}
															className="w-6 h-6"
														/>
														<span className="text-xs text-muted-foreground">
															{sender?.name}
														</span>
													</div>
												)}
												<div
													className={`p-3 rounded-lg ${
														isOwnMessage
															? 'bg-primary text-primary-foreground'
															: 'bg-muted text-foreground'
													}`}
												>
													{message.text && (
														<p className="text-sm">{message.text}</p>
													)}
													{message.image && (
														<img
															src={message.image}
															alt="Message attachment"
															className="mt-2 rounded max-w-full"
														/>
													)}
													<p
														className={`text-xs mt-1 ${
															isOwnMessage
																? 'text-primary-foreground/70'
																: 'text-muted-foreground'
														}`}
													>
														{new Date(
															message.createdAt
														).toLocaleTimeString([], {
															hour: '2-digit',
															minute: '2-digit'
														})}
													</p>
												</div>
											</div>
										</div>
									);
								})}
							</div>

							{/* Message Input */}
							<div className="p-4 border-t border-border bg-card">
								<div className="flex items-center space-x-2">
									<Button variant="ghost" size="sm">
										<Paperclip className="w-4 h-4" />
									</Button>
									<Button variant="ghost" size="sm">
										<ImageIcon className="w-4 h-4" />
									</Button>
									<div className="flex-1 relative">
										<Input
											placeholder="Type a message..."
											value={newMessage}
											onChange={(e) => setNewMessage(e.target.value)}
											onKeyPress={handleKeyPress}
											className="pr-12"
										/>
									</div>
									<Button
										onClick={handleSendMessage}
										disabled={!newMessage.trim()}
										size="sm"
									>
										<Send className="w-4 h-4" />
									</Button>
								</div>
							</div>
						</>
					) : (
						<div className="flex-1 flex items-center justify-center">
							<div className="text-center">
								<div className="w-16 h-16 mx-auto mb-4 bg-muted rounded-full flex items-center justify-center">
									<Search className="w-8 h-8 text-muted-foreground" />
								</div>
								<h3 className="text-lg font-medium text-foreground mb-2">
									Select a conversation
								</h3>
								<p className="text-muted-foreground">
									Choose a conversation from the list to start messaging
								</p>
							</div>
						</div>
					)}
				</div>
			</div>
		</div>
	);
}
