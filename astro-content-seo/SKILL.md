---
name: astro-content-seo
description: Publishing and SEO checklist for the user's Astro content sites — thebeautyanswer.com and cookingfix ("sister sites"), and any similar Astro + Tailwind + Preact content-collection site the user builds. Use this whenever adding or editing content entries (answers, recipes, articles), touching astro.config.mjs/content schema, generating images or social assets, or preparing a deploy — these sites follow a specific AI-content-generation + SEO-audit + image-generation pipeline that isn't obvious from a single file in isolation.
---

# Astro content-site publishing & SEO checklist

`thebeautyanswer.com` and `cookingfix` share the same pipeline: static Astro site,
Preact islands for interactivity, Tailwind for styling, content stored as Content
Collections with a Zod schema, PHP-backed admin/contact endpoints under `public/`.
New sites the user builds in this style will likely follow the same shape.

## Content model

Content lives under `src/content/<collection>/` as markdown/YAML frontmatter, validated
against a Zod schema in `src/content/config.ts`. Before adding a new entry, check that
schema for required vs optional fields (e.g. thebeautyanswer.com requires `title`,
`description`, `category`, `supplies`, `problems`, `fixes`, `tags`, `timeMinutes`) and
the category enum — content that doesn't validate will fail the build, not just look
wrong.

## AI content generation

New content is generated from a prompt template (`prompts/answer-generation.md` or
equivalent) via a script (`scripts/generate-answers.mjs` or equivalent), not written
by hand entry-by-entry. If asked to add a batch of new content, check for this script
first rather than hand-authoring frontmatter files one at a time — it's very likely
already wired to produce schema-valid entries.

## Before publishing new content, run:

1. **SEO audit** — `scripts/audit-seo.mjs` (or repo equivalent). Checks things like
   missing metadata, duplicate titles, broken internal links — run it, don't just
   eyeball the new page.
2. **Image pipeline** — new content typically needs:
   - `scripts/optimize-answer-images.mjs` (or equivalent) for on-page images
   - `scripts/generate-og-image.mjs` for the social share card
   - `scripts/generate-pins.mjs` / `pin-image.mjs` for Pinterest-format images, since
     these are recipe/beauty content sites where Pinterest is a real traffic source
3. **Build + preview** — `npm run build` then `npm run preview` before shipping;
   these are static sites, a broken build is a broken deploy, not a runtime surprise.

Check each site's actual `package.json` scripts and `scripts/` directory for the exact
filenames before running anything — names above are the pattern observed, not a
guarantee every site has all of them.

## Routing / config gotchas

- `trailingSlash: 'always'` is commonly set in `astro.config.mjs` — new internal links
  should match that convention or routing will break in production even if the dev
  server tolerates it.
- Site URL / base path often come from `ASTRO_SITE` / `ASTRO_BASE` env vars rather than
  being hardcoded — don't hardcode a domain in new components.
- Legacy slug redirects may be handled in the detail page itself (e.g. the
  `[slug].astro` route) rather than in a redirects config — check there before assuming
  a moved/renamed entry needs a separate redirect mechanism.

## Deploy

`deploy.sh` at the repo root, tag/branch-driven like the user's other projects. Check
it before assuming a generic `npm run build && rsync` flow — there's usually more to it
(cache busting, sitemap regen, etc.).

## Admin editor

There's usually a local-only admin/editor script (`npm run admin`) for editing content
without hand-writing frontmatter — prefer pointing the user at it over manual file
edits when the change is a content edit rather than a code change.
