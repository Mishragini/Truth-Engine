import ThemeToggle from "./theme";
import { Link } from "react-router";

export function Header() {
  return (
    <header className="border-b border-border/80 bg-background/95 backdrop-blur">
      <div className="content-container flex h-16 items-center justify-between">
        <Link className="text-sm font-semibold tracking-tight" to="/">
          Truth Engine
        </Link>
        <ThemeToggle />
      </div>
    </header>
  );
}
