import React from "react";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";

interface PageLayoutProps {
	title: string;
	subtitle?: string;
	actions?: React.ReactNode;
	children: React.ReactNode;
	onBack?: () => void;
	className?: string;
}

export function PageLayout({
	title,
	subtitle,
	actions,
	children,
	onBack,
	className,
}: PageLayoutProps) {
	return (
		<div className={`flex flex-col ${className || ""}`}>
			<div className="flex items-center justify-between px-6 py-4 shrink-0">
				<div className="flex items-center gap-4">
					{onBack && (
						<Button
							variant="ghost"
							size="icon-sm"
							onClick={onBack}
							className="text-muted-foreground hover:text-foreground -ml-2"
						>
							<ArrowLeft className="w-4 h-4" />
						</Button>
					)}
					<div>
						<h1 className="text-heading font-semibold tracking-[1.5px] uppercase text-foreground">
							{title}
						</h1>
						{subtitle && (
							<p className="text-ui text-muted-foreground mt-1">
								{subtitle}
							</p>
						)}
					</div>
				</div>
				{actions && <div className="flex items-center gap-2">{actions}</div>}
			</div>
			{children}
		</div>
	);
}
