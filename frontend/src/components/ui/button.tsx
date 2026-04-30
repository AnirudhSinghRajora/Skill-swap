import * as React from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';

import { cn } from '@/lib/utils';

const buttonVariants = cva(
	"inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-full text-sm font-medium tracking-tight transition-[transform,background-color,color,box-shadow,border-color] duration-200 ease-out disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:ring-[3px] focus-visible:ring-accent/60 focus-visible:ring-offset-2 focus-visible:ring-offset-background aria-invalid:ring-destructive/30 aria-invalid:border-destructive active:translate-y-px",
	{
		variants: {
			variant: {
				default:
					'bg-foreground text-background shadow-[0_1px_0_0_color-mix(in_oklab,var(--background)_30%,transparent)_inset,0_4px_14px_-6px_color-mix(in_oklab,var(--foreground)_55%,transparent)] hover:bg-foreground/92 active:shadow-[0_1px_0_0_color-mix(in_oklab,var(--background)_30%,transparent)_inset,0_2px_6px_-4px_color-mix(in_oklab,var(--foreground)_55%,transparent)]',
				brand:
					'bg-primary text-primary-foreground shadow-[0_1px_0_0_color-mix(in_oklab,#fff_22%,transparent)_inset,0_6px_18px_-8px_color-mix(in_oklab,var(--primary)_75%,transparent)] hover:bg-primary/92 active:shadow-[0_1px_0_0_color-mix(in_oklab,#fff_22%,transparent)_inset,0_2px_8px_-6px_color-mix(in_oklab,var(--primary)_75%,transparent)]',
				destructive:
					'bg-destructive text-destructive-foreground shadow-xs hover:bg-destructive/92 focus-visible:ring-destructive/30',
				outline:
					'border border-foreground/15 bg-transparent text-foreground hover:border-foreground/40 hover:bg-foreground/[0.03] dark:border-foreground/20 dark:hover:border-foreground/50',
				secondary:
					'bg-secondary text-secondary-foreground hover:bg-secondary/85',
				ghost: 'hover:bg-foreground/[0.05] hover:text-foreground dark:hover:bg-foreground/[0.08]',
				link: 'text-foreground underline-offset-4 decoration-accent decoration-2 hover:underline rounded-none'
			},
			size: {
				default: 'h-10 px-5 py-2 has-[>svg]:px-4',
				sm: 'h-8 px-3.5 gap-1.5 has-[>svg]:px-3',
				lg: 'h-12 px-7 text-[0.95rem] has-[>svg]:px-5',
				icon: 'size-10'
			}
		},
		defaultVariants: {
			variant: 'default',
			size: 'default'
		}
	}
);

function Button({
	className,
	variant,
	size,
	asChild = false,
	...props
}: React.ComponentProps<'button'> &
	VariantProps<typeof buttonVariants> & {
		asChild?: boolean;
	}) {
	const Comp = asChild ? Slot : 'button';

	return (
		<Comp
			data-slot="button"
			className={cn(buttonVariants({ variant, size, className }))}
			{...props}
		/>
	);
}

export { Button, buttonVariants };
