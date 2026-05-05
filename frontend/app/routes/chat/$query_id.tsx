import { useEffect, useRef, useState } from "react";
import { Link, useLocation, useParams } from "react-router";
import Markdown from "react-markdown";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";

export default function Page() {
  const { query_id } = useParams();
  const location = useLocation();
  const { cachedResponse, searchQuery } = location.state ?? {};

  const [finalContent, setFinalContent] = useState(cachedResponse);
  const [isStreaming, setIsStreaming] = useState(!cachedResponse);
  const [streamError, setStreamError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);
  // this is just a variable that survives re-renders
  // changing it does NOT cause a re-render
  const bufferRef = useRef("");

  // this is a reference to the actual <pre> element in the DOM
  // we use it to write text directly to the screen
  const outputRef = useRef<HTMLPreElement>(null);
  useEffect(() => {
    if (!query_id) {
      setStreamError("Missing query id.");
      setIsStreaming(false);
      return;
    }

    if (cachedResponse && refreshKey === 0) {
      setIsStreaming(false);
      return;
    }

    setFinalContent("");
    bufferRef.current = "";
    setStreamError(null);
    setIsStreaming(true);

    const evt_src = new EventSource(
      `${import.meta.env.VITE_BACKEND_BASE_URL}/search/${query_id}`,
    );
    evt_src.addEventListener("chunk", (e) => {
      // 1. add the new chunk to our buffer variable
      const parsed = JSON.parse(e.data);
      bufferRef.current += parsed;

      // 2. write the whole buffer directly to the screen
      // no React involved — just like vanilla JS
      if (outputRef.current) {
        outputRef.current.textContent = bufferRef.current;
      }
    });
    evt_src.addEventListener("full_response", (e) => {
      const parsed = JSON.parse(e.data);
      setFinalContent(parsed);
      setIsStreaming(false);

      evt_src.close();
    });
    evt_src.addEventListener("error", (e) => {
      if (e instanceof MessageEvent) {
        const parsed = JSON.parse(e.data);
        setStreamError(parsed.message);
        toast.error(parsed.message);
      }
    });
    evt_src.onerror = () => {
      setStreamError("Connection lost. Please retry.");
      setIsStreaming(false);
      toast.error("Connection lost");
      evt_src.close();
    };
    return () => evt_src.close();
  }, [cachedResponse, query_id, refreshKey]);

  const showStreaming = isStreaming && !finalContent;
  const showError = Boolean(streamError && !finalContent);

  return (
    <div className="page-shell">
      <section className="content-container flex w-full flex-col gap-6 py-10">
        <div className="flex items-center justify-between gap-3 border-b border-border pb-4">
          <div>
            <p className="text-sm text-muted-foreground">Query</p>
            <h1 className="page-heading">
              {searchQuery ?? `Result ${query_id ?? ""}`}
            </h1>
          </div>
          <div className="flex items-center gap-2">
            <Button asChild variant="outline">
              <Link to="/">New search</Link>
            </Button>
            <Button
              onClick={() => {
                setRefreshKey((value) => value + 1);
              }}
              variant="ghost"
            >
              Retry
            </Button>
          </div>
        </div>

        {showStreaming && (
          <div className="rounded-xl border border-border bg-card p-5">
            <p className="mb-3 text-sm text-muted-foreground">
              Generating response...
            </p>
            <pre
              className="min-h-28 whitespace-pre-wrap wrap-break-word font-mono text-sm"
              ref={outputRef}
            />
          </div>
        )}

        {showError && (
          <div className="rounded-xl border border-destructive/30 bg-destructive/10 p-4">
            <p className="text-sm text-destructive">{streamError}</p>
          </div>
        )}

        {finalContent && (
          <article className="markdown-content rounded-xl border border-border bg-card p-6">
            <Markdown>{finalContent}</Markdown>
          </article>
        )}
      </section>
    </div>
  );
}
