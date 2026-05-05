import { useMutation } from "@tanstack/react-query";
import { Button } from "./ui/button";
import { ButtonGroup } from "./ui/button-group";
import { Field, FieldDescription, FieldLabel } from "./ui/field";
import { Input } from "./ui/input";
import { performSearch } from "~/lib/api";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { useNavigate } from "react-router";

export function SearchComponent() {
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const { mutate, isPending } = useMutation({
    mutationKey: ["search"],
    mutationFn: (data: string) => performSearch(data),
    onSuccess: (result) => {
      const { query_id, response } = result.data;
      navigate(`/chat/${query_id}`, {
        state: { cachedResponse: response, searchQuery: query },
      });
    },
    onError: (error) => {
      let error_message =
        error instanceof Error ? error.message : "Failed to perform search";
      toast.error(error_message);
    },
  });
  const trimmedQuery = useMemo(() => query.trim(), [query]);

  return (
    <form
      className="w-full max-w-2xl"
      onSubmit={(event) => {
        event.preventDefault();
        if (!trimmedQuery) return;
        mutate(trimmedQuery);
      }}
    >
      <Field>
        <FieldLabel htmlFor="input-button-group">Search</FieldLabel>
        <FieldDescription>
          Ask one clear question and get a streamed answer.
        </FieldDescription>
        <ButtonGroup>
          <Input
            value={query}
            disabled={isPending}
            onChange={(e) => {
              setQuery(e.target.value);
            }}
            id="input-button-group"
            placeholder="Type to search..."
          />
          <Button disabled={isPending || !trimmedQuery} type="submit">
            {isPending ? "Searching..." : "Search"}
          </Button>
        </ButtonGroup>
      </Field>
    </form>
  );
}
