import { Titlebar } from "./titlebar";

interface AppLayoutProps {
	children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
	return (
		<div className="h-screen flex flex-col bg-background text-foreground overflow-hidden">
			<Titlebar />
			<main className="flex-1 overflow-y-auto">
				{children}
			</main>
		</div>
	);
}
