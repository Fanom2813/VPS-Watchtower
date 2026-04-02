import { Titlebar } from "./titlebar";

interface AppLayoutProps {
	children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
	return (
		<div className="h-screen flex flex-col bg-background text-foreground">
			<Titlebar />
			<main className="flex-1 flex flex-col min-h-0 overflow-hidden">
				{children}
			</main>
		</div>
	);
}
