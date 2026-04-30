'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useEditor, EditorContent } from '@tiptap/react';
import { mergeAttributes } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import LinkExtension from '@tiptap/extension-link';
import ImageExtension from '@tiptap/extension-image';
import Placeholder from '@tiptap/extension-placeholder';
import imageCompression from 'browser-image-compression';
import {
  Bold,
  Italic,
  Strikethrough,
  Code,
  Link as LinkIcon,
  List,
  ListOrdered,
  ImagePlus,
  Send,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import api from '@/lib/api';
import { encryptBinary } from '@/lib/e2ee/crypto';

interface ChatInputProps {
  conversationId?: string;
  onSend: (content: string, imageIds: string[]) => void;
  onTyping?: () => void;
  disabled?: boolean;
  placeholder?: string;
  sharedKey?: Uint8Array | null;
}

const MAX_IMAGE_SIZE = 10 * 1024 * 1024; // 10 MB before compression
const DRAFT_KEY_PREFIX = 'skillswap:chat-draft:';

const EncryptedImageExtension = ImageExtension.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      encryptedImageId: {
        default: null,
        parseHTML: (element: HTMLElement) => element.getAttribute('data-encrypted-image-id'),
        renderHTML: (attributes: { encryptedImageId?: string | null }) => {
          if (!attributes.encryptedImageId) return {};
          return { 'data-encrypted-image-id': attributes.encryptedImageId };
        },
      },
    };
  },
  renderHTML({ HTMLAttributes }) {
    const attrs = { ...HTMLAttributes } as Record<string, string | null | undefined>;
    if (attrs['data-encrypted-image-id']) {
      delete attrs.src;
    }
    return ['img', mergeAttributes(this.options.HTMLAttributes, attrs)];
  },
});

const QUICK_STARTERS = [
  'Can we lock in a 30-minute swap this week?',
  'I can teach this in exchange for help with that. Does that work for you?',
  'Want to jump on a video call to align on goals?',
];

export function ChatInput({
  conversationId,
  onSend,
  onTyping,
  disabled = false,
  placeholder = 'Type a message…',
  sharedKey,
}: ChatInputProps) {
  const [uploading, setUploading] = useState(false);
  const [, setRenderTrigger] = useState(0);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const onTypingRef = useRef(onTyping);
  onTypingRef.current = onTyping;
  const handleSendRef = useRef<() => void>(() => {});
  const draftStorageKey = conversationId ? `${DRAFT_KEY_PREFIX}${conversationId}` : null;
  const draftStorageKeyRef = useRef<string | null>(draftStorageKey);
  draftStorageKeyRef.current = draftStorageKey;

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
      }),
      LinkExtension.configure({
        openOnClick: false,
        HTMLAttributes: {
          target: '_blank',
          rel: 'noopener noreferrer',
        },
      }),
      EncryptedImageExtension.configure({
        inline: false,
        allowBase64: false,
      }),
      Placeholder.configure({ placeholder }),
    ],
    editorProps: {
      handleKeyDown(_view, event) {
        if (event.key === 'Enter' && !event.shiftKey && !event.ctrlKey && !event.metaKey) {
          event.preventDefault();
          handleSendRef.current();
          return true;
        }
        return false;
      },
    },
    onUpdate({ editor: currentEditor }) {
      onTypingRef.current?.();

      if (typeof window === 'undefined') return;

      const key = draftStorageKeyRef.current;
      if (!key) return;

      const html = currentEditor.getHTML();
      const hasText = currentEditor.getText().trim().length > 0;
      const hasImages = html.includes('<img');

      if (hasText || hasImages) {
        window.localStorage.setItem(key, html);
      } else {
        window.localStorage.removeItem(key);
      }
    },
    onTransaction() {
      setRenderTrigger((c) => c + 1);
    },
  });

  // Sync editable state with disabled prop
  useEffect(() => {
    editor?.setEditable(!disabled);
  }, [editor, disabled]);

  // Restore draft when opening a conversation
  useEffect(() => {
    if (!editor || !draftStorageKey || typeof window === 'undefined') return;

    const savedDraft = window.localStorage.getItem(draftStorageKey);
    if (!savedDraft) {
      editor.commands.clearContent();
      return;
    }

    editor.commands.setContent(savedDraft, { emitUpdate: false });
  }, [editor, draftStorageKey]);

  // ── Send ──────────────────────────────────────────────────────────────────

  const handleSend = useCallback(() => {
    if (!editor) return;
    const html = editor.getHTML();
    const hasText = editor.getText().trim().length > 0;
    const hasImages = html.includes('<img');
    if (!hasText && !hasImages) return;

    // Extract encrypted image IDs from data attributes
    const imageIds: string[] = [];
    const imgIdRe = /data-encrypted-image-id="([^"]+)"/g;
    let m: RegExpExecArray | null;
    while ((m = imgIdRe.exec(html)) !== null) {
      imageIds.push(m[1]);
    }

    onSend(html, imageIds);
    editor.commands.clearContent();

    if (draftStorageKeyRef.current && typeof window !== 'undefined') {
      window.localStorage.removeItem(draftStorageKeyRef.current);
    }
  }, [editor, onSend]);

  handleSendRef.current = handleSend;

  // ── Image upload ──────────────────────────────────────────────────────────

  // Keep sharedKey in a ref so the upload callback always sees the latest value
  const sharedKeyRef = useRef(sharedKey);
  sharedKeyRef.current = sharedKey;

  const handleImageUpload = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file || !editor) return;
      e.target.value = '';

      if (!file.type.startsWith('image/')) {
        toast.error('Only image files are supported');
        return;
      }
      if (file.size > MAX_IMAGE_SIZE) {
        toast.error('Image too large (max 10 MB)');
        return;
      }

      setUploading(true);
      try {
        const compressed = await imageCompression(file, {
          maxSizeMB: 1,
          maxWidthOrHeight: 1920,
          useWebWorker: true,
        });

        const key = sharedKeyRef.current;

        if (!key) {
          toast.error('Encryption is required. Wait until secure chat is ready.');
          return;
        }

        // Encrypted upload: encrypt the compressed bytes, upload as blob
        const plainBytes = new Uint8Array(await compressed.arrayBuffer());
        const encryptedBytes = encryptBinary(plainBytes, key);
        const encryptedBlob = new Blob([encryptedBytes], { type: 'application/octet-stream' });

        const res = await api.chatImages.upload(encryptedBlob, true);
        // Insert as an actual TipTap image node. Raw insertContent can be serialized
        // as text in some cases when src is omitted.
        editor
          .chain()
          .focus()
          .setImage({
            src: 'encrypted-image',
            alt: 'Encrypted image',
            encryptedImageId: res.image_id,
          } as unknown as { src: string; alt: string; encryptedImageId: string })
          .run();
      } catch {
        toast.error('Failed to upload image');
      } finally {
        setUploading(false);
      }
    },
    [editor],
  );

  // ── Link toggle ───────────────────────────────────────────────────────────

  const toggleLink = useCallback(() => {
    if (!editor) return;
    if (editor.isActive('link')) {
      editor.chain().focus().unsetLink().run();
      return;
    }
    const href = window.prompt('Enter URL:');
    if (!href) return;
    const url = href.match(/^https?:\/\//) ? href : `https://${href}`;
    try {
      new URL(url);
    } catch {
      toast.error('Invalid URL');
      return;
    }
    editor.chain().focus().setLink({ href: url }).run();
  }, [editor]);

  const insertQuickStarter = useCallback(
    (text: string) => {
      editor?.chain().focus().insertContent(`${text} `).run();
    },
    [editor],
  );

  // ── Render ────────────────────────────────────────────────────────────────

  const canSend =
    !!editor &&
    (editor.getText().trim().length > 0 || editor.getHTML().includes('<img'));

  return (
    <div className="border-t border-border bg-card">
      {/* Toolbar */}
      <div role="toolbar" aria-label="Text formatting" className="flex flex-wrap items-center gap-0.5 border-b border-border/50 px-3 py-1.5">
        <ToolbarBtn
          onClick={() => editor?.chain().focus().toggleBold().run()}
          active={editor?.isActive('bold')}
          disabled={disabled}
          title="Bold"
        >
          <Bold className="h-3.5 w-3.5" />
        </ToolbarBtn>
        <ToolbarBtn
          onClick={() => editor?.chain().focus().toggleItalic().run()}
          active={editor?.isActive('italic')}
          disabled={disabled}
          title="Italic"
        >
          <Italic className="h-3.5 w-3.5" />
        </ToolbarBtn>
        <ToolbarBtn
          onClick={() => editor?.chain().focus().toggleStrike().run()}
          active={editor?.isActive('strike')}
          disabled={disabled}
          title="Strikethrough"
        >
          <Strikethrough className="h-3.5 w-3.5" />
        </ToolbarBtn>
        <ToolbarBtn
          onClick={() => editor?.chain().focus().toggleCode().run()}
          active={editor?.isActive('code')}
          disabled={disabled}
          title="Inline code"
        >
          <Code className="h-3.5 w-3.5" />
        </ToolbarBtn>

        <Divider />

        <ToolbarBtn
          onClick={toggleLink}
          active={editor?.isActive('link')}
          disabled={disabled}
          title="Link"
        >
          <LinkIcon className="h-3.5 w-3.5" />
        </ToolbarBtn>
        <ToolbarBtn
          onClick={() => editor?.chain().focus().toggleBulletList().run()}
          active={editor?.isActive('bulletList')}
          disabled={disabled}
          title="Bullet list"
        >
          <List className="h-3.5 w-3.5" />
        </ToolbarBtn>
        <ToolbarBtn
          onClick={() => editor?.chain().focus().toggleOrderedList().run()}
          active={editor?.isActive('orderedList')}
          disabled={disabled}
          title="Ordered list"
        >
          <ListOrdered className="h-3.5 w-3.5" />
        </ToolbarBtn>

        <Divider />

        <ToolbarBtn
          onClick={() => fileInputRef.current?.click()}
          disabled={disabled || uploading}
          title="Upload image"
        >
          {uploading ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <ImagePlus className="h-3.5 w-3.5" />
          )}
        </ToolbarBtn>

        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleImageUpload}
          className="hidden"
          tabIndex={-1}
        />

        {/* Send hint */}
        <span className="ml-auto text-[10px] text-muted-foreground select-none">
          Enter to send · Shift+Enter for new line
        </span>
      </div>

      {/* Quick starters */}
      {!!editor && editor.isEmpty && !disabled && (
        <div className="flex flex-wrap items-center gap-2 px-4 pt-3">
          {QUICK_STARTERS.map((starter) => (
            <button
              key={starter}
              type="button"
              onClick={() => insertQuickStarter(starter)}
              className="rounded-full border border-border/80 bg-background px-3 py-1 text-xs text-muted-foreground transition-colors hover:border-primary/40 hover:text-foreground"
            >
              {starter}
            </button>
          ))}
        </div>
      )}

      {/* Editor + send button */}
      <div className="flex items-end gap-2 px-4 py-3">
        <div className="flex-1 min-h-[2.5rem] max-h-[12rem] overflow-y-auto rounded-lg border border-border bg-background px-3 py-2 text-sm focus-within:ring-2 focus-within:ring-primary/40">
          <EditorContent editor={editor} />
        </div>
        <Button
          size="icon"
          onClick={handleSend}
          disabled={disabled || !canSend}
          aria-label="Send message (Enter)"
          title="Send (Enter)"
        >
          <Send className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

// ── Sub-components ───────────────────────────────────────────────────────────

function ToolbarBtn({
  onClick,
  active,
  disabled,
  title,
  children,
}: {
  onClick: () => void;
  active?: boolean;
  disabled?: boolean;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={title}
      aria-label={title}
      className={cn(
        'inline-flex items-center justify-center rounded p-1.5 text-muted-foreground transition-colors',
        'hover:bg-accent hover:text-foreground',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-1',
        'disabled:pointer-events-none disabled:opacity-50',
        active && 'bg-accent text-foreground',
      )}
    >
      {children}
    </button>
  );
}

function Divider() {
  return <div className="mx-1 h-4 w-px bg-border" aria-hidden />;
}
