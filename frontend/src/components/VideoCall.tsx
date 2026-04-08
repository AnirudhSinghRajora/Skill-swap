"use client";

import { useCallback, useState } from "react";
import {
  LiveKitRoom,
  VideoConference,
  RoomAudioRenderer,
} from "@livekit/components-react";
import "@livekit/components-styles";
import { Button } from "@/components/ui/button";
import { Phone, PhoneOff, X } from "lucide-react";
import api from "@/lib/api";

interface VideoCallProps {
  conversationId: string;
  onClose: () => void;
}

export function VideoCall({ conversationId, onClose }: VideoCallProps) {
  const [token, setToken] = useState<string | null>(null);
  const [url, setUrl] = useState<string>("");
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const joinCall = useCallback(async () => {
    setConnecting(true);
    setError(null);
    try {
      const data = await api.video.getToken(conversationId);
      setToken(data.token);
      setUrl(data.url);
    } catch {
      setError("Failed to join call. Video calls may not be configured.");
    } finally {
      setConnecting(false);
    }
  }, [conversationId]);

  const handleDisconnect = useCallback(() => {
    setToken(null);
    setUrl("");
    onClose();
  }, [onClose]);

  // Not yet connected — show join button
  if (!token) {
    return (
      <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-border bg-card p-8">
        <Phone className="h-12 w-12 text-indigo-500" />
        <h3 className="text-lg font-semibold">Video Call</h3>
        {error && (
          <p className="text-sm text-destructive text-center max-w-xs">{error}</p>
        )}
        <div className="flex gap-3">
          <Button onClick={joinCall} disabled={connecting}>
            {connecting ? "Connecting..." : "Join Call"}
          </Button>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
        </div>
      </div>
    );
  }

  // Connected — show LiveKit video conference
  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-black">
      <div className="flex items-center justify-between bg-zinc-900 px-4 py-2">
        <span className="text-sm font-medium text-white">Video Call</span>
        <Button
          variant="ghost"
          size="icon"
          className="text-white hover:bg-zinc-800"
          onClick={handleDisconnect}
        >
          <X className="h-5 w-5" />
        </Button>
      </div>

      <div className="flex-1">
        <LiveKitRoom
          serverUrl={url}
          token={token}
          connect={true}
          onDisconnected={handleDisconnect}
          data-lk-theme="default"
          style={{ height: "100%" }}
        >
          <VideoConference />
          <RoomAudioRenderer />
        </LiveKitRoom>
      </div>

      <div className="flex justify-center bg-zinc-900 py-3">
        <Button
          variant="destructive"
          size="lg"
          className="rounded-full"
          onClick={handleDisconnect}
        >
          <PhoneOff className="mr-2 h-5 w-5" />
          End Call
        </Button>
      </div>
    </div>
  );
}
