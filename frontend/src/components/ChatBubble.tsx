'use client';

import { useCallback, useMemo, useState } from 'react';
import DOMPurify from 'dompurify';
import { cn } from '@/lib/utils';
import { Lock } from 'lucide-react';
import { ImageLightbox } from '@/components/ImageLightbox';
import { EncryptedImage } from '@/components/EncryptedImage';
import type { MessageType } from '@/types/chat';

// ── DOMPurify configuration ──────────────────────────────────────────────────

const SANITIZE_CONFIG = {
  ALLOWED_TAGS: [
    'p', 'br', 'strong', 'b', 'em', 'i', 's', 'del',
    'a', 'code', 'pre', 'blockquote',
    'ul', 'ol', 'li',
    'h1', 'h2', 'h3',
    'img', 'span',
  ],
  ALLOWED_ATTR: ['href', 'target', 'rel', 'src', 'alt', 'class', 'data-encrypted-image-id'],
};

const UUID_RE = '[0-9a-fA-F-]{36}';
const ENCRYPTED_TAG_ESCAPED_DOUBLE = new RegExp(
  `&lt;img[^&]*data-encrypted-image-id=(?:&quot;|")(${UUID_RE})(?:&quot;|")[^&]*\\/?&gt;`,
  'gi',
);
const ENCRYPTED_TAG_ESCAPED_SINGLE = new RegExp(
  `&lt;img[^&]*data-encrypted-image-id=(?:&#39;|')(${UUID_RE})(?:&#39;|')[^&]*\\/?&gt;`,
  'gi',
);

/**
 * Ensure links open in a new tab and images lazy-load.
 * Hooks are additive, so guard with a module-level flag.
 */
let hooksRegistered = false;
function ensureDOMPurifyHooks() {
  if (hooksRegistered || typeof window === 'undefined') return;
  DOMPurify.addHook('afterSanitizeAttributes', (node) => {
    if (node.tagName === 'A') {
      node.setAttribute('target', '_blank');
      node.setAttribute('rel', 'noopener noreferrer');
    }
    if (node.tagName === 'IMG') {
      node.setAttribute('loading', 'lazy');
    }
  });
  hooksRegistered = true;
}

/** True when the string contains at least one HTML element. */
const HAS_HTML = /<[a-z][\s\S]*?>/i;

/**
 * Some editor serialization paths can produce escaped encrypted image tags
 * (`&lt;img ... data-encrypted-image-id=... /&gt;`) as plain text.
 * Normalize only this known-safe placeholder shape back to an image tag
 * before sanitization.
 */
function normalizeEncryptedImagePlaceholders(content: string): string {
  return content
    .replace(
      ENCRYPTED_TAG_ESCAPED_DOUBLE,
      '<img alt="Encrypted image" data-encrypted-image-id="$1" />',
    )
    .replace(
      ENCRYPTED_TAG_ESCAPED_SINGLE,
      '<img alt="Encrypted image" data-encrypted-image-id="$1" />',
    );
}

/**
 * Sanitise message content for safe rendering.
 * Plain-text messages (no HTML tags) get newlines converted to `<br>`.
 */
function sanitizeContent(content: string): string {
  const normalized = normalizeEncryptedImagePlaceholders(content);

  if (typeof window === 'undefined') {
    // SSR: escape everything – client hydration will apply proper sanitization
    return normalized
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\n/g, '<br>');
  }

  ensureDOMPurifyHooks();

  if (!HAS_HTML.test(normalized)) {
    // Legacy plain-text message – escape & convert newlines
    const escaped = normalized
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\n/g, '<br>');
    return DOMPurify.sanitize(escaped, SANITIZE_CONFIG);
  }

  return DOMPurify.sanitize(normalized, SANITIZE_CONFIG);
}

// ── Component ────────────────────────────────────────────────────────────────

interface ChatBubbleProps {
  message: MessageType;
  isOwn: boolean;
  sharedKey?: Uint8Array | null;
}

/** Regex to detect encrypted-image placeholder <img> tags in sanitized HTML. */
const ENCRYPTED_IMG_RE = new RegExp(
  `<img[^>]*data-encrypted-image-id=["'](${UUID_RE})["'][^>]*\\/?>`,
  'gi',
);

/**
 * Split sanitized HTML into alternating segments of plain HTML and encrypted
 * image IDs so we can render React components for the encrypted images.
 */
function splitEncryptedImages(html: string): Array<{ type: 'html'; value: string } | { type: 'encrypted-image'; imageId: string }> {
  const segments: Array<{ type: 'html'; value: string } | { type: 'encrypted-image'; imageId: string }> = [];
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  const re = new RegExp(ENCRYPTED_IMG_RE.source, 'g');
  while ((match = re.exec(html)) !== null) {
    if (match.index > lastIndex) {
      segments.push({ type: 'html', value: html.slice(lastIndex, match.index) });
    }
    segments.push({ type: 'encrypted-image', imageId: match[1] });
    lastIndex = re.lastIndex;
  }
  if (lastIndex < html.length) {
    segments.push({ type: 'html', value: html.slice(lastIndex) });
  }
  return segments;
}

export function ChatBubble({ message, isOwn, sharedKey }: ChatBubbleProps) {
  const html = message._decryptionFailed ? '' : sanitizeContent(message.content);
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

  const segments = useMemo(() => splitEncryptedImages(html), [html]);
  const hasEncryptedImages = segments.some((s) => s.type === 'encrypted-image');

  const handleContentClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      const target = e.target as HTMLElement;
      if (target.tagName === 'IMG') {
        e.preventDefault();
        const src = (target as HTMLImageElement).src;
        if (src) setLightboxSrc(src);
      }
    },
    [],
  );

  return (
    <div
      className={cn('flex w-full', isOwn ? 'justify-end' : 'justify-start')}
    >
      <div
        className={cn(
          'max-w-[75%] rounded-2xl px-4 py-2 text-sm leading-relaxed break-words',
          isOwn
            ? 'bg-primary text-primary-foreground rounded-br-md'
            : 'bg-muted text-foreground rounded-bl-md',
        )}
      >
        {message._decryptionFailed ? (
          <div className="flex items-center gap-1.5 italic opacity-60">
            <Lock className="h-3.5 w-3.5 shrink-0" />
            <span>Unable to decrypt this message</span>
          </div>
        ) : hasEncryptedImages ? (
          <div className="chat-content" onClick={handleContentClick}>
            {segments.map((seg, i) =>
              seg.type === 'html' ? (
                <span
                  key={i}
                  dangerouslySetInnerHTML={{ __html: seg.value }}
                  suppressHydrationWarning
                />
              ) : (
                <EncryptedImage
                  key={seg.imageId}
                  imageId={seg.imageId}
                  sharedKey={sharedKey ?? null}
                  onClick={(blobUrl) => setLightboxSrc(blobUrl)}
                />
              ),
            )}
          </div>
        ) : (
          <div
            className="chat-content"
            dangerouslySetInnerHTML={{ __html: html }}
            suppressHydrationWarning
            onClick={handleContentClick}
          />
        )}

        {/* Meta line: time + edited badge + encrypted indicator */}
        <div
          className={cn(
            'mt-1 flex items-center gap-1.5 text-[10px]',
            isOwn
              ? 'text-primary-foreground/60 justify-end'
              : 'text-muted-foreground justify-start',
          )}
        >
          {message.encrypted && !message._decryptionFailed && (
            <Lock className="h-2.5 w-2.5" />
          )}
          {message.is_edited && <span>edited</span>}
          <time dateTime={message.created_at}>{formatMessageTime(message.created_at)}</time>
        </div>
      </div>

      {lightboxSrc && (
        <ImageLightbox
          src={lightboxSrc}
          onClose={() => setLightboxSrc(null)}
        />
      )}
    </div>
  );
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatMessageTime(iso: string): string {
  const date = new Date(iso);
  if (isNaN(date.getTime())) return '';
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
