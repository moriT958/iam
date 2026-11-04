import type { APIRoute } from "astro";
import { getCollection } from "astro:content";

// Manifest of the posts that have their own page, read by the Go server at
// startup so that it accepts likes only for posts that really exist.
// The filter must match getStaticPaths() in posts/[id].astro.
export const GET: APIRoute = async () => {
  const posts = await getCollection("posts");
  const slugs = posts
    .filter((p) => (import.meta.env.DEV || !p.data.draft) && !p.data.url)
    .map((p) => p.id)
    .sort();

  return new Response(JSON.stringify({ slugs }), {
    headers: { "Content-Type": "application/json" },
  });
};
