---
name: veteran-search
description: "Use when a question needs facts from the web, a code host, or a large source tree and a first keyword search came back thin — search like an old internet hand (Gopher, Archie, Usenet, FTP indexes): find the primary source, browse its tree, search exact strings, grep local mirrors, follow the citations, and record where each fact came from."
version: 1.0.0
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [search, research, github, grep, provenance]
    category: research
---

# Search like someone who learned on Gopher

People who found things before web search did it by knowing where things live. Archie indexed file
names on FTP servers, Veronica indexed Gopher menus, and you read the FAQ before you asked. The same
habits beat a pile of keyword searches today. A search engine returns pages that talk about a
thing. These steps get you the thing.

## 1. Go to the primary source, not the write-up

Ask first: who produces this data? A game client, a vendor's API docs, a repo, a standards body,
a dataset. Blog posts and guides are pointers to that. Cite the source, not the guide.

- Code and docs: the project's repo (README, `docs/`, `CHANGELOG`, releases, issues, discussions).
- Game or product facts: the client's own files and the vendor's notes, then community measurements.
- A guide is useful for finding names to search for, and weak as evidence.

## 2. Read the directory before you guess queries

Gopher was menus. List the tree, then pick.

```bash
gh api 'repos/OWNER/REPO/git/trees/HEAD?recursive=1' --jq '.tree[] | select(.type=="blob") | .path'
gh api repos/OWNER/REPO/readme --jq .content | base64 -d | less
gh api repos/OWNER/REPO/branches --jq '.[].name'        # other branches often hold what you need
gh release view TAG -R OWNER/REPO --json assets
```

Folder names and file names tell you more than a first query would. Read the README and FAQ, then
the file that names your topic. Look at the branch list: a `forever`, `beta` or `next` branch may
be the one that matters.

## 3. Search exact strings, with operators

Search for the thing the source will literally contain: a function name, an error message, an
event name, a config key, a version string. Quote it.

- Web: `"exact phrase"`, `site:github.com`, `filetype:lua`, `inurl:`, a minus term to drop noise.
- Code host: `gh search code '"Name = \"GetItemStatDelta\"" repo:OWNER/REPO'`.
- Add the vendor's own words ("LiteralName", "Payload", "SecretReturns") rather than your paraphrase.

## 4. Mirror it and grep locally

Anything you will query more than twice belongs on disk. A local grep is exact, fast and repeatable.

```bash
git clone --depth 1 URL            # or fetch just the branch you need
git grep -n 'PATTERN' origin/BRANCH -- path/      # search a branch without checking it out
git log -S'identifier' --oneline -- path/          # when did this appear or vanish
git show origin/BRANCH:path/file | sed -n 100,140p
gh api -H "Accept: application/vnd.github.raw" repos/OWNER/REPO/contents/PATH
```

Prefer the host's API or raw files over scraping rendered pages. Download data dumps (JSON, CSV)
and query them with `jq` or Python.

## 5. Follow the citations

Good sources point at better ones. Read "Related work", "Credits", "Sources", changelogs, linked
issues and pull requests, the repos a project forks or depends on, and the generated-from list in
its docs. One link in a README led to the upstream data repo; the same trick finds the dataset a
tool was built from.

## 6. Triangulate and date everything

- Two independent sources agreeing beats one source repeated.
- A measurement (probe output, a dump from the real client) beats a claim.
- Write down the build, commit, release or date next to each fact. Facts about a beta rot.
- Different sources disagree: say so, and say which you trust and why.

## 7. Absence is evidence, with care

If the primary source's API list lacks a name, it probably does not exist there. Say "not found in
X at COMMIT", not "does not exist". Check the other branch and one alternate spelling first.

## 8. Be economical

- A standard web search first; widen to a thorough search when results are thin, off target or old.
- Send independent searches together.
- Stop when a primary source answers the question.

## 9. Report like a librarian

Give the answer, then the evidence: file and line, URL, commit, build, or probe output. Separate
what you verified, what a source claims, and what you infer. Name licences when you reuse
anything, and read the licence of anything you copy.

## Example: one session, applied

Question: what does WoW Forever's client offer for specs and talents? A search for guides gave
priorities but no numbers. Listing a GitHub repo's tree led to a stat-weights file and its credits.
Another repo's branch list showed a `forever` branch of Blizzard's UI source already cloned
locally; `git grep` over that branch settled which functions exist. An in-game probe then
confirmed what a level 22 character really returns. Each fact carried its source.
