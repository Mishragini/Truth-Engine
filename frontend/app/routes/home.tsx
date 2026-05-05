import { SearchComponent } from "~/components/search";
import type { Route } from "./+types/home";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Truth Engine" },
    { name: "description", content: "Welcome to Truth Engine!" },
  ];
}

export default function Home() {
  return (
    <div className="page-shell">
      <div className="content-container flex min-h-[calc(100vh-4rem)] flex-col items-center justify-center gap-8 py-10">
        <div className="space-y-3 text-center">
          <h1 className="page-heading">Search with clarity</h1>
          <p className="max-w-xl text-sm text-muted-foreground md:text-base">
            Enter your question and get a concise, streamed response.
          </p>
        </div>
        <SearchComponent />
      </div>
    </div>
  );
}
