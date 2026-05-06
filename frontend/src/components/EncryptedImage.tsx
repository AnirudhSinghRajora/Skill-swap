'use client';

import { useEffect, useRef, useState } from 'react';
import { Lock, ImageOff, Loader2 } from 'lucide-react';
import { decryptBinary } from '@/lib/e2ee/crypto';
import api from '@/lib/api';

interface EncryptedImageProps {
  imageId: string;
  sharedKey: Uint8Array | null;
  onClick?: (blobUrl: string) => void;
  className?: string;
}

/**
 * Fetches an encrypted image from the server, decrypts it client-side,
 * and renders it as a blob: URL. Revokes the URL on unmount to avoid leaks.
 */
export function EncryptedImage({ imageId, sharedKey, onClick, className }: EncryptedImageProps) {
  const [blobUrl, setBlobUrl] = useState<string | null>(null);
  const [error, setError] = useState(false);
  const [loading, setLoading] = useState(true);
  const blobUrlRef = useRef<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function fetchAndDecrypt() {
      if (!sharedKey) {
        setError(true);
        setLoading(false);
        return;
      }

      try {
        const url = api.chatImages.getUrl(imageId);
        const res = await fetch(url);
        if (!res.ok) throw new Error('fetch failed');

        const encryptedBytes = new Uint8Array(await res.arrayBuffer());
        const plainBytes = decryptBinary(encryptedBytes, sharedKey);

        if (cancelled) return;

        if (!plainBytes) {
          setError(true);
          setLoading(false);
          return;
        }

        const blob = new Blob([plainBytes]);
        const objectUrl = URL.createObjectURL(blob);
        blobUrlRef.current = objectUrl;
        setBlobUrl(objectUrl);
      } catch {
        if (!cancelled) setError(true);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    fetchAndDecrypt();

    return () => {
      cancelled = true;
      if (blobUrlRef.current) {
        URL.revokeObjectURL(blobUrlRef.current);
        blobUrlRef.current = null;
      }
    };
  }, [imageId, sharedKey]);

  if (loading) {
    return (
      <div className="inline-flex items-center justify-center rounded-lg bg-muted/50 p-6">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !blobUrl) {
    return (
      <div className="inline-flex items-center gap-1.5 rounded-lg bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
        {sharedKey ? (
          <>
            <ImageOff className="h-3.5 w-3.5" />
            <span>Unable to decrypt image</span>
          </>
        ) : (
          <>
            <Lock className="h-3.5 w-3.5" />
            <span>Encrypted image</span>
          </>
        )}
      </div>
    );
  }

  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img
      src={blobUrl}
      alt="Encrypted image"
      loading="lazy"
      className={className ?? 'max-w-full cursor-pointer rounded-lg'}
      onClick={() => onClick?.(blobUrl)}
    />
  );
}
