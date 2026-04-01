import { Minus, Square, X } from "lucide-react";
import { getPlatform } from "@/lib/platform";

const platform = getPlatform();

function WindowControlsMac() {
	return (
		<div className="electrobun-webkit-app-region-no-drag flex items-center gap-[6px] px-3">
			<button
				onClick={() => window.dispatchEvent(new CustomEvent("window:close"))}
				className="group w-3 h-3 rounded-full bg-smui-surface-3 hover:bg-[#ff5f57] flex items-center justify-center transition-colors"
			>
				<X className="w-1.5 h-1.5 text-transparent group-hover:text-[#4a0002] transition-colors" />
			</button>
			<button
				onClick={() => window.dispatchEvent(new CustomEvent("window:minimize"))}
				className="group w-3 h-3 rounded-full bg-smui-surface-3 hover:bg-[#febc2e] flex items-center justify-center transition-colors"
			>
				<Minus className="w-1.5 h-1.5 text-transparent group-hover:text-[#5f4a00] transition-colors" />
			</button>
			<button
				onClick={() => window.dispatchEvent(new CustomEvent("window:maximize"))}
				className="group w-3 h-3 rounded-full bg-smui-surface-3 hover:bg-[#28c840] flex items-center justify-center transition-colors"
			>
				<Square className="w-1.5 h-1.5 text-transparent group-hover:text-[#006500] transition-colors" />
			</button>
		</div>
	);
}

function WindowControlsRight() {
	return (
		<div className="electrobun-webkit-app-region-no-drag flex items-center h-full">
			<button
				onClick={() => window.dispatchEvent(new CustomEvent("window:minimize"))}
				className="w-8 h-full flex items-center justify-center text-muted-foreground/50 hover:text-muted-foreground transition-colors"
			>
				<Minus className="w-2.5 h-2.5" />
			</button>
			<button
				onClick={() => window.dispatchEvent(new CustomEvent("window:maximize"))}
				className="w-8 h-full flex items-center justify-center text-muted-foreground/50 hover:text-muted-foreground transition-colors"
			>
				<Square className="w-2 h-2" />
			</button>
			<button
				onClick={() => window.dispatchEvent(new CustomEvent("window:close"))}
				className="w-8 h-full flex items-center justify-center text-muted-foreground/50 hover:text-[hsl(var(--smui-red))] transition-colors"
			>
				<X className="w-2.5 h-2.5" />
			</button>
		</div>
	);
}

export function Titlebar() {
	const isMac = platform === "mac";

	return (
		<div className="electrobun-webkit-app-region-drag h-8 flex items-center border-b border-border bg-smui-surface-1 select-none shrink-0">
			{isMac && <WindowControlsMac />}

			<span className="text-label tracking-[2px] uppercase text-muted-foreground pl-3 flex-1">
				eyes on vps
			</span>

			{!isMac && <WindowControlsRight />}
		</div>
	);
}
