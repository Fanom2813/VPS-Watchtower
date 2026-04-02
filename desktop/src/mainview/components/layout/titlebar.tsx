import { Minus, X, Settings } from "lucide-react";
import { useNavigate } from "react-router";
import { getPlatform } from "@/lib/platform";
import { rpc } from "@/lib/rpc";

const platform = getPlatform();

function WindowControlsMac() {
	const handleClose = () => {
		rpc.request.closeWindow({});
	};

	const handleMinimize = () => {
		rpc.request.minimizeWindow({});
	};

	return (
		<div className="electrobun-webkit-app-region-no-drag flex items-center gap-[6px] px-3">
			<button
				onClick={handleClose}
				className="group w-3 h-3 rounded-full bg-smui-surface-3 hover:bg-[#ff5f57] flex items-center justify-center transition-colors"
			>
				<X className="w-1.5 h-1.5 text-transparent group-hover:text-[#4a0002] transition-colors" />
			</button>
			<button
				onClick={handleMinimize}
				className="group w-3 h-3 rounded-full bg-smui-surface-3 hover:bg-[#febc2e] flex items-center justify-center transition-colors"
			>
				<Minus className="w-1.5 h-1.5 text-transparent group-hover:text-[#5f4a00] transition-colors" />
			</button>
		</div>
	);
}

function WindowControlsRight() {
	const handleClose = () => {
		rpc.request.closeWindow({});
	};

	const handleMinimize = () => {
		rpc.request.minimizeWindow({});
	};

	return (
		<div className="electrobun-webkit-app-region-no-drag flex items-center h-full">
			<button
				onClick={handleMinimize}
				className="w-8 h-full flex items-center justify-center text-muted-foreground/50 hover:text-muted-foreground transition-colors"
			>
				<Minus className="w-2.5 h-2.5" />
			</button>
			<button
				onClick={handleClose}
				className="w-8 h-full flex items-center justify-center text-muted-foreground/50 hover:text-[hsl(var(--smui-red))] transition-colors"
			>
				<X className="w-2.5 h-2.5" />
			</button>
		</div>
	);
}

export function Titlebar() {
	const isMac = platform === "mac";
	const navigate = useNavigate();

	return (
		<div className="electrobun-webkit-app-region-drag h-8 flex items-center border-b border-border bg-smui-surface-1 select-none shrink-0">
			{isMac && <WindowControlsMac />}

			<span className="text-label tracking-[2px] uppercase text-muted-foreground pl-3 flex-1">
				eyes on vps
			</span>

			<button
				className="electrobun-webkit-app-region-no-drag w-8 h-full flex items-center justify-center text-muted-foreground/50 hover:text-foreground transition-colors"
				title="Settings"
				onClick={() => navigate("/settings")}
			>
				<Settings className="w-3.5 h-3.5" />
			</button>

			{!isMac && <WindowControlsRight />}
		</div>
	);
}
