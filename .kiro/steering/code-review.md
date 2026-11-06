---
inclusion: always
---

# Code Review Discipline (Genesys Cloud Terraform Provider)

Apply this whenever you are asked to review code, a diff, or a pull request for
this repository. The goal is an evidence-first review that is neither biased
toward approval nor toward nit-picking, and that does not miss the obvious.

## The one rule that matters most

**No claim without a citation.** Every review comment is either:

- **Verified** — you read the relevant code and can cite `file:line` that proves
  the point. State it as a finding.
- **Unverified** — you have not confirmed it against the code. State it as a
  *question to the author*, never as a defect.

If you cannot point to the code that proves a concern is real, it is a question,
not a finding. Do not label anything "blocking" without a citation.

- **An anchor must sit on a line this PR changed.** If the concern is about code
  the diff does not touch (a sibling path that wasn't updated, a pre-existing gap),
  it is an out-of-scope *question*, not a finding — raise it as "was leaving X out
  intentional?" and anchor it to a line the PR *did* change (or to none). Never pin
  a comment to an unchanged line as though the author failed to edit it there.

## Investigate before you judge (order of operations)

Gather all evidence first, then write the review. Do not produce a verdict in the
same pass as the investigation — that is how anchoring bias happens.

For every changed field, function, or API call, read the code it depends on
before commenting:

1. **Schema** — `resource_genesyscloud_*_schema.go`. Check type, `Optional` /
   `Required` / `Computed`, `ForceNew`, `ValidateFunc`, defaults, `StateFunc`,
   `MaxItems`. Ask: is the "bad" input you are worried about even reachable?
2. **SDK / struct model** — follow the `.amazonq/rules/sdk-reference.md` rule and
   read the real SDK types from `.sdk-reference/`. Never guess field names, types,
   or nullability. Read the local `*_struct.go` too — including `MarshalJSON`.
   The struct tag (`omitempty`) is not the truth; the marshaller is.
3. **Proxy / API layer** — `genesyscloud_*_proxy.go`. Determine whether the API
   call is a **full-representation replace** or a **partial merge/patch**. This
   single question changes how you judge any "we now send null / omit this field"
   change. Read the actual path and HTTP method.
4. **CRUD parity** — if a behavior changed in one of create/read/update/delete,
   check the other three for consistency (and the exporter `getAll`).

## Verify assumptions that commonly produce false positives

Before flagging behavior as a bug, confirm:

- **Is the input reachable?** A `ValidateFunc` may make the value you are worried
  about impossible (e.g. `IntAtLeast(60)` means `0` only ever means "unset").
- **Full-replace vs merge PATCH?** In a full-replace update, sending `null` for a
  field absent from config is correct, not a silent wipe — Terraform config is
  declarative and complete on every apply. A full-representation PUT (e.g.
  `PutRoutingQueue`, `PutOutboundContactlist`) submits the *entire* object, so "we
  now send the whole slice" or "we omit this on unset" is usually correct behavior,
  not a data loss. Confirm the method is a PUT of the full object before you judge
  such a change a wipe.
- **What does serialization actually emit?** Read `MarshalJSON` / `SetField`
  semantics, not the JSON tag.
- **Does `HasChange` / `GetOk` behave as you assume?** `GetOk` reports the type's
  zero value as absent. `HasChange` can be defeated by a `StateFunc` that
  normalizes values — removing it may be a fix, not a regression.
- **Is this a defect or a design preference?** A different structure, a parameter
  you'd have passed explicitly instead of via context, a naming choice — these are
  preferences, not bugs, when the code works. Raise them as ❓ or 🟢 at most, and
  never list a preference among the changes that "need a fix". Severity inflation
  (flagging preferences as blocking) is the most common over-reach to guard against.
- **Is a "the repo always does X" claim actually checked across the repo?** Before
  contrasting the change with "the convention" or "everywhere else," grep for the
  other occurrences and read them. One example is not a convention. If you have
  confirmed only one comparison point, name it ("unlike `FetchPagesConcurrently`")
  rather than generalizing ("unlike everywhere else"). A scoped, cited contrast
  beats a sweeping one you cannot back.

## Completeness sweep (catch the obvious you would otherwise miss)

Bias is not only false positives; it is tunnel vision on the interesting change
while missing the plain gap. Before finishing, sweep for:

- **Error and nil paths** — nil dereferences, ignored errors, unchecked type
  assertions, missing `resp` nil-guards.
- **Deleted-code guarantees** — when a change removes a function or branch,
  enumerate the guarantees the old code provided (nil-guards, state cleanup,
  error handling) and confirm each is still preserved or provably unnecessary.
  A regression often hides in what was deleted, not in what was added. Example:
  a resolver that guarded `resource.State != nil` is removed and the surviving
  path now derefs `resource.State.ID` unconditionally.
- **Reachability of the changed block** — trace which control-flow paths actually
  reach the changed lines (which `switch` case, which early-return guards precede
  it, what type the value has there). A block that reads as general may only be
  correct for one type/branch — e.g. a var-substitution that overwrites a value
  with a string is only valid for scalar attributes, not nested maps/arrays.
- **CRUD + import + exporter parity** — does read populate everything create/update
  set? Does import work with only the resource ID? Does the exporter cover new refs?
- **Retries / concurrency** — retryable vs non-retryable classification, 400/409
  handling, `WithRetries` timeouts.
- **State drift** — will `terraform plan` settle after apply, or does read write
  back a value the config never set (perpetual diff)?
- **Test coverage of the changed lines** — do the new/changed branches actually
  have a test? A green suite that never exercises the new branch proves nothing.
- **Repo conventions** — `.amazonq/rules/*.md`, `DICTIONARY.md` domain language,
  package/file layout in `new-resource.md`, generated `docs/` must not be hand-edited.

## Output format

Group comments by severity and cite evidence for each:

- **🔴 Blocking** — correctness/security/data-loss, verified with `file:line`.
- **🟡 Should-fix** — maintainability, missing tests, convention violations.
- **🟢 Nit** — style, naming, comments.
- **❓ Question** — unverified assumption; ask the author, do not assert.

End with an explicit recommendation: Approve / Approve-with-nits / Request-changes,
and list the specific conditions that must be met to reach approval.

## Anchor every comment (so it can be pasted on the PR)

The reviewer places each comment at an exact spot in the diff. Every comment must
lead with an anchor:

```
📍 <relative/path/to/file.go>:<line>   ` <verbatim code at that line> `
```

- Use the line number in the **current branch version** of the file, not the diff
  hunk offset. Re-read the line in the file to confirm the number before finalizing.
- Quote the line (or the key line of a range `Lstart–Lend`) verbatim so the anchor
  survives small line-number drift.

## Write like a human, not a report

Comments must read like an experienced teammate on GitHub, not a linter or an AI:

- Lead with a layman explanation, then the technical context. Say what breaks and
  why it matters in plain terms a non-specialist could follow, then attach the
  precise technical reason. Don't open with jargon or scanner-speak, and never give
  only the jargon — the comment should be something any teammate can act on without
  decoding it. Explain the *why* in 1–4 sentences.
- First person and honest questions when unsure ("Am I right that…?").
- Vary phrasing; do not template every comment. No `Issue:/Impact:/Fix:` labels
  inside prose, no severity emoji in the sentence, no "As an AI" / "It is worth
  noting" / "Overall, the code" throat-clearing.
- Acknowledge genuinely good work briefly; be specific, not effusive.
- Suggest a fix when useful, but avoid pasting large code blocks unless asked.

## Self-check before posting

- Did I read schema + struct/marshaller + proxy for every field I commented on?
- Is every "blocking" backed by a citation, or is it actually a question?
- Did I run the completeness sweep, or only review the interesting hunk?
- Am I stating a design decision as a bug without confirming intent?
- Does every anchor land on a line this PR actually changed? (Out-of-diff concerns
  are questions, not findings.)
- Every "the repo always/never does X" claim — did I grep the other cases, or
  generalize from one?
