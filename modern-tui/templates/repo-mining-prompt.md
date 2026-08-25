# Codex Prompt — Mine Existing TUI Repositories

Inspect the TUI repositories available in this workspace and extract the owner's repeated UI/UX and implementation patterns into the existing `modern-tui` skill.

Do not redesign the applications and do not blindly copy project-specific code into the skill.

## Goals

1. Identify patterns repeated across at least two TUI projects.
2. Distinguish deliberate house style from one-off implementation details.
3. Capture reusable design rules, architecture rules, interaction conventions, and proven component patterns.
4. Add concise examples only when they materially help another coding agent reproduce the pattern.
5. Keep `SKILL.md` lean; place detail in `references/`.

## Repositories to inspect

Prioritize any available repositories matching or related to:

- Tide
- TideMail
- tideui
- TideFTP
- WhatTheDock / TideDock or other Docker TUI projects
- other terminal applications by the same owner

## Specifically inspect

- layout modes and responsive fallbacks
- pane focus and selected-row styling
- border strategy
- status/footer hint conventions
- key routing and focus traversal
- search and help behavior
- command palette implementations
- modal/dialog visual language
- theme definitions and contrast handling
- compact vs comfortable density
- terminal resize handling
- bounded rendering
- scrolling helpers
- async/background task patterns
- empty/loading/error states
- confirmation flows
- sparklines, meters, and one-line graphs
- mouse behavior if present
- tests for state transitions and rendering

## Rules for updating the skill

- Preserve the existing `name: modern-tui`.
- Do not add instructions that apply to only one application unless clearly labeled as a Tide-family convention.
- Do not prescribe Rust or Go universally. Preserve the framework already used by a project.
- Favor semantic principles over exact pixel/character constants unless the constant is a deliberate reusable design token.
- Avoid duplicate rules across files.
- Keep reference files one level below `SKILL.md` when possible.
- Add a new reference file only when the topic is substantial enough to justify it.
- Do not add secrets, private endpoints, credentials, or machine-specific paths.

## Deliverable

Update the skill in place, then report:

1. repositories inspected
2. recurring patterns found
3. skill files changed
4. patterns deliberately rejected as too project-specific
5. any contradictions between projects that need a design decision
6. validation performed
