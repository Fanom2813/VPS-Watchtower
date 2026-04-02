import { useState } from "react";
import { useNavigate } from "react-router";
import { Button } from "@/components/ui/button";
import { Shield, ArrowRight, Loader2, ClipboardPaste } from "lucide-react";
import { useAgentsStore } from "@/stores/agents";
import { rpc } from "@/lib/rpc";
import { PageLayout } from "@/components/layout/page-layout";

export function AuthPage() {
	const navigate = useNavigate();
	const { agents, addAgent } = useAgentsStore();
	const [url, setUrl] = useState("");
	const [token, setToken] = useState("");
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(false);

	const handleConnect = async () => {
		if (!url || !token) return;

		setError("");
		setLoading(true);

		try {
			await addAgent(url, token);
			navigate("/overview");
		} catch (err) {
			setError(err instanceof Error ? err.message : "Connection failed");
		} finally {
			setLoading(false);
		}
	};

	const hasAgents = agents.length > 0;

	const content = (
		<div className="w-full max-w-sm space-y-6">
			<div className="text-center space-y-2">
				<Shield className="w-8 h-8 text-primary mx-auto" />
				<h1 className="text-heading font-semibold tracking-[1.5px] uppercase text-foreground">
					{hasAgents ? "Connect Agent" : "Add Your First Server"}
				</h1>
				<p className="text-ui text-muted-foreground">
					Enter the agent URL and pairing token from your VPS
				</p>
			</div>

			<div className="space-y-4">
				<div className="space-y-2">
					<label className="text-label text-muted-foreground tracking-[1.5px] uppercase block">
						Agent URL
					</label>
					<div className="relative">
						<input
							type="text"
							value={url}
							onChange={(e) => setUrl(e.target.value)}
							placeholder="ws://your-vps-ip:9090/ws"
							className="w-full border border-border bg-card px-3 py-2 pr-10 text-ui text-foreground font-mono placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-primary"
						/>
						<button
							onClick={async () => {
								const text = await rpc.request.readClipboard({});
								setUrl(text.trim());
							}}
							className="absolute right-0 top-0 h-full px-3 text-muted-foreground hover:text-foreground transition-colors"
							title="Paste"
						>
							<ClipboardPaste className="w-3.5 h-3.5" />
						</button>
					</div>
				</div>

				<div className="space-y-2">
					<label className="text-label text-muted-foreground tracking-[1.5px] uppercase block">
						Pairing Token
					</label>
					<div className="relative">
						<input
							type="text"
							value={token}
							onChange={(e) => setToken(e.target.value)}
							placeholder="Token from eyes-agent setup"
							className="w-full border border-border bg-card px-3 py-2 pr-10 text-ui text-foreground font-mono placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-primary"
						/>
						<button
							onClick={async () => {
								const text = await rpc.request.readClipboard({});
								setToken(text.trim());
							}}
							className="absolute right-0 top-0 h-full px-3 text-muted-foreground hover:text-foreground transition-colors"
							title="Paste"
						>
							<ClipboardPaste className="w-3.5 h-3.5" />
						</button>
					</div>
				</div>

				{error && (
					<p className="text-label text-destructive tracking-wider">
						{error}
					</p>
				)}
			</div>

			<Button
				onClick={handleConnect}
				disabled={!url || !token || loading}
				className="w-full uppercase tracking-wider text-label"
			>
				{loading ? (
					<>
						<Loader2 className="w-3.5 h-3.5 mr-2 animate-spin" />
						Connecting...
					</>
				) : (
					<>
						Connect
						<ArrowRight className="w-3.5 h-3.5 ml-2" />
					</>
				)}
			</Button>

			<p className="text-label text-muted-foreground/60 tracking-wider text-center">
				Run{" "}
				<span className="text-smui-frost-2 font-mono">
					eyes-agent setup --port 9090
				</span>{" "}
				on your VPS to get the token
			</p>
		</div>
	);

	if (!hasAgents) {
		return (
			<div className="h-full flex items-center justify-center px-6">
				{content}
			</div>
		);
	}

	return (
		<PageLayout
			title="Add Server"
			onBack={() => navigate("/overview")}
			className="h-full"
		>
			<div className="h-full flex items-center justify-center px-6">
				{content}
			</div>
		</PageLayout>
	);
}
