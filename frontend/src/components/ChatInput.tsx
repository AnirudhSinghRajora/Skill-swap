'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useEditor, EditorContent } from '@tiptap/react';
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
  onSend: (content: string, imageIds: string[]) => void;
  onTyping?: () => void;
  disabled?: boolean;
  placeholder?: string;
  sharedKey?: Uint8Array | null;
}

const MAX_IMAGE_SIZE = 10 * 1024 * 1024; // 10 MB before compression

export function ChatInput({
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
      ImageExtension.configure({
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
    onUpdate() {
      onTypingRef.current?.();
    },
    onTransaction() {
      setRenderTrigger((c) => c + 1);
    },
  });

  // Sync editable state with disabled prop
  useEffect(() => {
    editor?.setEditable(!disabled);
  }, [editor, disabled]);

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

        if (key) {
          // Encrypted upload: encrypt the compressed bytes, upload as blob
          const plainBytes = new Uint8Array(await compressed.arrayBuffer());
          const encryptedBytes = encryptBinary(plainBytes, key);
          const encryptedBlob = new Blob([encryptedBytes], { type: 'application/octet-stream' });

          const res = await api.chatImages.upload(encryptedBlob, true);
          // Insert a placeholder image with a data attribute instead of a src URL
          editor
            .chain()
            .focus()
            .setImage({ src: '', alt: 'Encrypted image' })
            .run();

          // TipTap doesn't natively support data-* attributes on images,
          // so we set it on the last inserted <img> via the DOM
          const view = editor.view;
          const imgs = view.dom.querySelectorAll('img[alt="Encrypted image"]:not([data-encrypted-image-id])');
          const lastImg = imgs[imgs.length - 1];
          if (lastImg) {
            lastImg.setAttribute('data-encrypted-image-id', res.image_id);
            lastImg.removeAttribute('src');
          }
        } else {
          // Plaintext upload: existing behavior
          const res = await api.chatImages.upload(compressed);
          const url = api.chatImages.getUrl(res.image_id);
          editor.chain().focus().setImage({ src: url }).run();
        }
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

  // ── Render ────────────────────────────────────────────────────────────────

  const isEmpty = !editor || editor.isEmpty;

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

      {/* Editor + send button */}
      <div className="flex items-end gap-2 px-4 py-3">
        <div className="flex-1 min-h-[2.5rem] max-h-[12rem] overflow-y-auto rounded-lg border border-border bg-background px-3 py-2 text-sm focus-within:ring-2 focus-within:ring-primary/40">
          <EditorContent editor={editor} />
        </div>
        <Button
          size="icon"
          onClick={handleSend}
          disabled={disabled || isEmpty}
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
