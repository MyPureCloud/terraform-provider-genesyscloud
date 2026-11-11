---
name: pr-review
description: >-
  Perform a rigorous, evidence-first pull request review for the Genesys Cloud
  Terraform provider. Use when the user asks to review a PR, review a branch,
  review code changes, or give review feedback / a review comment before approval.
  Produces citation-backed findings grouped by severity plus an explicit
  approve / request-changes recommendation.
---

# PR Review Skill

A structured workflow for reviewing changes to the terraform-provider-genesyscloud
repo. Optimized to avoid two failure modes: (1) asserting a bug before verifying it
against ground truth, and (2) tunnel-visioning on the interesting change and missing
an obvious gap. Read `.kiro/steering/code-review.md` for the underlying discipline;
this skill is the step-by-step procedure.

## Step 1 — Establish scope (gather, do not judge)

1. Identify the branch and the correct base to diff against. The local `main` is
   often stale; diff against the actual merge base or the PR's parent commit, not a
   blanket `main...HEAD`, or the file list will be polluted with unrelated commits.
   - `git merge-base main HEAD` to find the true base, or diff against the parent of
     the first branch commit.
2. Get the changed-files list and the PR description (if a URL/number is given).
   Prefer `gh` if it's installed (common on macOS/Linux) — it gives the PR, its
   diff, and the base directly. If `gh` isn't available (often the case on
   Windows), don't keep retrying it: fetch the PR web page for the title/description
   and derive the diff base from git locally. If the page can't be reached either,
   ask the user to paste the description.
3. **Reconcile the PR number against the branch before scoping.** The requested PR
   number, the current branch name, and the HEAD commit's message/label do not
   always agree (a branch named `DEVTOOLING-1764` has shipped a commit labeled
   `pr-2520` while the user asked about `#2519`). If they disagree, state the
   discrepancy and confirm which one you're reviewing rather than guessing.
4. Separate real changes from noise (e.g. regenerated `docs/` from an unrelated
   doc-regen commit). State which files are in scope.
5. **Review from the current file, not the raw diff.** Use the diff only to get the
   changed-file list, then read the *current branch version* of each file with the
   file tools — that's what Step 2 needs for accurate line numbers anyway. On
   Windows/PowerShell this is doubly important: `git diff` output is frequently
   UTF-16 encoded and the terminal echoes/garbles long output, so the raw diff is
   unreliable to read directly there.

## Step 2 — Read ground truth for every touched field/function

Do NOT comment yet. For each changed field, function, or API interaction, read:

- **Schema** (`*_schema.go`): type, Optional/Required/Computed, ForceNew,
  ValidateFunc, defaults, StateFunc, MaxItems.
- **Struct + marshaller** (`*_struct.go`) and the **SDK model**. Read `MarshalJSON` /
  `SetField` — the JSON tag is not the truth. Reach the SDK model via
  `.sdk-reference/` per `.amazonq/rules/sdk-reference.md` (the symlink in that rule
  works on macOS/Linux). If `.sdk-reference/` is absent — typical on Windows, where
  the symlink command doesn't apply — read the model straight from the Go module
  cache: get the SDK version from `go.mod`, then read it under
  `<GOPATH>/pkg/mod/github.com/mypurecloud/platform-client-sdk-go/...@<version>/platformclientv2/`.
  Resolve `<GOPATH>` with `go env GOPATH` (`$(go env GOPATH)` on macOS/Linux,
  `$env:GOPATH` or the same `go env GOPATH` on Windows PowerShell).
- **Proxy** (`*_proxy.go`): the exact API path + HTTP method. Decide
  full-replace vs partial-merge PATCH — this determines whether "send null / omit"
  changes are correct.
- **The other CRUD funcs + exporter** for parity with the changed one.

## Step 3 — Convert concerns into verified findings or questions

For each concern, write it as a hypothesis with a disproof condition, then go find
the code that confirms or refutes it. Only after checking:

- If the code proves it → **finding**, cite `file:line`.
- If the code refutes it → drop it (and note it if you nearly flagged it).
- If you cannot tell from the code → **question** to the author, not a defect.

Common false positives to rule out first:
- Is the "bad" input reachable given the ValidateFunc/schema?
- Is this a full-replace update (null-on-unset is correct) or a merge patch?
- Does `GetOk` zero-value / `StateFunc` normalization explain the code as-is?

## Step 4 — Completeness sweep

Independently of the diff's focus, check: nil/error paths, unchecked type
assertions, CRUD+import+exporter parity, retry classification (400/409/404),
`terraform plan` settling (no perpetual diff from read writing back unset values),
and whether the changed branches are actually covered by a test.

Two sweeps that are easy to skip and have bitten past reviews:

- **What the deleted code guaranteed.** When the PR removes a function or branch,
  list the guarantees it provided (nil-guards, state cleanup, error handling) and
  confirm each still holds on the surviving path. The regression is often in what
  was removed, not what was added.
- **Whether the changed block is actually reached the way you assume.** Trace the
  control flow to the edited lines — which `switch` case, which guards precede it,
  what type the value has there. A block that looks general may only be correct
  for one type (e.g. a scalar), and silently wrong for nested maps/arrays.
- **Out-of-diff siblings are scope questions, not findings.** If a parallel code
  path wasn't updated (e.g. an MRMO/alternate exporter that didn't get the new
  flag), confirm it's actually outside this diff before writing it up — then raise
  it as "is that follow-up intended?" anchored to a line the PR changed, not as a
  defect pinned to the untouched file.

## Step 5 — Write the review

Group comments by severity: 🔴 Blocking · 🟡 Should-fix · 🟢 Nit · ❓ Question.
Every comment must be **anchored** and must **sound like a human wrote it**.

### Anchor every comment to a file and line

The reviewer needs to paste each comment at an exact spot on the PR. For every
comment provide:

- **File path** relative to repo root.
- **Line number** (or a range `Lstart–Lend`) in the *current* branch version of
  the file, not the diff hunk offset. Get it from the file you actually read.
- **The line's content** quoted verbatim (one line, or the key line of a range)
  so the reviewer can confirm the anchor even if line numbers shift.

Verify each line number by re-reading that line in the file before you finalize —
do not estimate from the diff, and never trust a line number from a panic stack
trace or a compiled binary. Those drift from current HEAD (a trace once pointed at
line 1093 while the real source line was 1075). The anchor number must come from
the file you just read. Use this header on each comment:

```
📍 <relative/path/to/file.go>:<line>   ` <verbatim code at that line> `
```

### Make it read like a human review comment

Write the way an experienced teammate leaves a GitHub review — not a linter, not a
report generator. Concretely:

- **Lead with a layman explanation, then the technical context.** Open with what
  breaks and why it matters in plain terms a non-specialist could follow, then
  attach the precise technical reason for the engineer who needs it. Don't open
  with jargon or scanner-speak. Prefer "this crashes the provider whenever the API
  response leaves out `worktype`" first, and *then* "because `Worktype` is a pointer
  and it's dereferenced here without a nil check." Never give only the jargon
  ("Potential nil-pointer dereference vulnerability detected") — a comment should be
  something any teammate can act on without having to decode it.
- Use natural first-person and questions where you are genuinely unsure:
  "Am I right that the delete PATCH can never legitimately need to clear this?"
- Keep it short. One to four sentences per comment. Explain the *why* and, when
  useful, suggest the fix — but do not paste giant code blocks unless asked.
- Vary the phrasing. Do not start every comment with the same word or template
  sentence. No emoji-severity inside the prose, no "Issue:/Impact:/Fix:" labels.
- Be collegial and specific. Acknowledge good work where it is real, briefly.
- No AI throat-clearing ("As an AI…", "It is worth noting that…", "Overall, the
  code…"). Just say the thing a senior engineer would say.

### Per-comment shape

```
📍 genesyscloud/.../resource_x.go:178   ` _ = d.Set("status_id", *ws.Worktype.Id+"/"+*ws.Id) `

`Worktype` is a pointer in the SDK model and this deref will panic the provider if
a GET ever comes back without the worktype expansion. Probably rare in practice,
but since you're already rewriting these lines, mind adding a nil guard?
```

Close with an explicit recommendation (Approve / Approve-with-nits /
Request-changes) and the exact conditions required to approve. Keep the summary a
few sentences — it should read like a person signing off, not a scorecard.

## Step 5.5 — Triage your own comments before delivering

The first pass tends to over-produce — the recurring correction in practice is
severity inflation, not missed bugs. Before presenting, re-pass every 🔴 and 🟡 and
ask: *would this survive the author pushing back?*

- If it's a **design preference** (idiom, structure, "I'd rather it took a param"),
  it is not a defect. Downgrade it to ❓ or 🟢, or drop it — never ship a preference
  as a 🔴/🟡 "fix".
- If the code **refutes** it once you look again, drop it (note it if you nearly
  flagged it).
- Keep 🔴/🟡 only for things you can defend with `file:line`.

Then offer two views so the author can act fast:

- **Fix-only list** — just the real defects (🔴 plus 🟡 that are genuine bugs or
  missing-test gaps). This is what "give me the comments that need only a fix"
  asks for; lead with it.
- **Full list** — everything, including nits and questions, for completeness.

## Step 6 — Self-check before delivering

- Every comment has a `file:line` anchor with the verbatim line, and I re-read the
  file to confirm the line number is current (not a diff offset).
- Every 🔴 has a citation (else it is a ❓).
- Read schema + marshaller + proxy for each commented field.
- Ran the completeness sweep, not just the headline change.
- Did not assert a deliberate design decision as a bug.
- Triaged the comments (Step 5.5): no design preference is sitting in the fix-only
  list, and every 🔴/🟡 would survive the author pushing back.
- Comments read like a human teammate: plain, varied phrasing, no report-style
  labels or AI throat-clearing.
- Every anchor is on a line the PR changed; out-of-diff and "whole-repo
  convention" claims were each either verified across their full scope or
  downgraded to a scoped question.
