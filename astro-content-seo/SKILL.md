---
name: astro-content-seo
description: Use when adding/editing content, touching the content schema, generating images, or deploying on the user's Astro content sites (thebeautyanswer.com, cookingfix) — these share one AI-generation + SEO-audit + image pipeline that a single file doesn't reveal.
version: 1.0.0
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [astro, seo, content, tailwind, preact, static-site, pinterest]
    category: productivity
---

# Astro content-site publishing & SEO checklist

`thebeautyanswer.com` and `cookingfix` share the same pipeline: static Astro site,
Preact islands for interactivity, Tailwind for styling, content stored as Content
Collections with a Zod schema, PHP-backed admin/contact endpoints under `public/`.
New sites the user builds in this style will likely follow the same shape.

Neither site is necessarily cloned on the current machine. Both are **private** repos
under `allisonhere` (`thebeautyanswer.com`, `cookingfix`) — inspect without cloning:

```bash
gh api repos/allisonhere/thebeautyanswer.com/contents/scripts --jq '.[].name'
gh api repos/allisonhere/cookingfix/contents/package.json --jq '.content' | base64 -d | jq .scripts
```


## Content model

Content lives under `src/content/<collection>/` as markdown/YAML frontmatter, validated
against a Zod schema in `src/content/config.ts`. Before adding a new entry, check that
schema for required vs optional fields (e.g. thebeautyanswer.com requires `title`,
`description`, `category`, `supplies`, `problems`, `fixes`, `tags`, `timeMinutes`) and
the category enum — content that doesn't validate will fail the build, not just look
wrong.

## AI content generation

New content is generated from a prompt template via a script, not hand-authored entry
by entry. On `thebeautyanswer.com` that is `scripts/generate-answers.mjs`
(`npm run gen:answers`); `cookingfix` does not currently ship a generator script, so
recipe entries there are added another way — check before assuming symmetry.

If asked for a batch of new content, look for the generator first:

```bash
jq -r '.scripts | to_entries[] | "\(.key): \(.value)"' package.json
ls scripts/
```

## Before publishing new content, run:

Verified script inventory (re-check with `ls scripts/` — sites drift):

| Step | thebeautyanswer.com | cookingfix |
|------|---------------------|------------|
| SEO audit | `npm run audit:seo` (`scripts/audit-seo.mjs`) | `npm run audit:seo` |
| Hub audit | `npm run audit:hubs` (`scripts/audit-problem-hubs.mjs`) | — not present |
| On-page images | `npm run optimize:images` → `scripts/optimize-answer-images.mjs --update-content --archive-originals` | `scripts/optimize-recipe-images.mjs` |
| Social card | `npm run gen:og` (`scripts/generate-og-image.mjs`) | same |
| Pinterest images | `npm run gen:pins` (`scripts/generate-pins.mjs`, `scripts/pin-image.mjs`) | same |
| Image sourcing | `npm run fetch:images` (`scripts/fetch-images.js`) | same |
| Slug maintenance | `scripts/add-recipe-slugs.mjs` | same |

Then build and preview:

```bash
npm run build     # runs postbuild: node scripts/stamp-offline-canonical.mjs
npm run preview
```

`postbuild` is load-bearing on `thebeautyanswer.com` — it stamps offline canonicals, so
never substitute a bare `astro build` for `npm run build`. `cookingfix` has no
`stamp-offline-canonical.mjs`; confirm with `jq -r '.scripts.postbuild' package.json`
before assuming either way.

Run the audits — don't eyeball the new page. A broken build is a broken deploy on a
static site, not a runtime surprise.

`npm run dev:all` (`scripts/dev-all.mjs`) is the multi-process dev entry point; prefer it
over `npm run dev` when the admin/PHP endpoints matter.

Pinterest matters for both sites (beauty/recipe content), which is why the pin scripts
exist separately from the OG card — don't skip them for content meant to be shareable.


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

`npm run admin` (`node admin/server.mjs`) is a local-only editor for content entries.
Prefer pointing the user at it for content edits; reserve hand-editing frontmatter for
code/schema changes.
