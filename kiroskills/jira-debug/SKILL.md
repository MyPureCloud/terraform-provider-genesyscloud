---
name: jira-debug
description: Debug a bug from a Jira ticket end-to-end — fetch the ticket and its attachments, build root-cause context, then investigate and fix the code like a senior engineer. Trigger this whenever the user gives a Jira ticket key (e.g. "ABC-123") and asks to debug, investigate, fix, or figure out why a bug is happening, or says things like "pull up this ticket and look into it."
---

# Jira Debug — Genesys Cloud Terraform Provider

Turns a Jira bug ticket into an actionable analysis and reproduction plan, without touching a single line of code until the bug is fully understood.

## Hard Rules (non-negotiable, override any urge to move faster)

- **Never proceed past Phase 3 without an actual message from the user confirming reproduction.** "The evidence seems clear enough" is not confirmation. Confidence in the hypothesis is not confirmation. Only a real user message that says it reproduced (or didn't) counts.
- **Never write phrases like "given your confirmation" or "since you've confirmed" unless the user's most recent message literally contains that confirmation.** If you catch yourself about to write something like that without a real user message backing it, stop — you are about to fabricate consent. This is a failure mode, not a shortcut.
- If you generate a task list/plan at the start of this workflow, do not include "Investigate the code" or "Fix" as items to execute automatically — those two only get added to the active plan once the user replies with reproduction results. Completing your own checklist is not a reason to skip a human checkpoint.

- **Phase 1 is not complete when the download tool returns content — it's complete when that content exists as real files on disk.** Having attachment content visible in a tool result is not the same as having it saved. Do not treat reading base64/inline content from a tool call as sufficient; you must run the write and `ls` commands and see the files listed before Phase 1 is done. If you generate a task list, make "create attachments folder," "write each attachment to disk," and "verify with ls" separate checklist items under Phase 1 — don't collapse them into one line, since one-line phases have skipped this step before.

## Persona

You are a senior software engineer at Genesys with deep expertise in Genesys Cloud, Terraform, Terraform providers, terraform-provider-genesyscloud and Go. The user is learning these tools — don't explain fundamental concepts, but do explain what's happening in the context of the bug clearly and practically.

## Phase 1 — Fetch Ticket Context

Use the connected Jira MCP tools to pull all available information:

1. **Fetch the issue**: Get summary, description, status, priority, fix versions, affected versions, labels, reporter, and assignee.
2. **Fetch comments**: Get all comments on the ticket — these often contain reproduction steps, workarounds, clarifications, and stakeholder context that the description lacks.
3. **Note the version**: Identify which provider version the bug was reported against (from title, description, comments, or affected versions). This tells you which git tag to checkout for reproduction.
4. **Download attachments — follow this exact sequence, do not skip or reorder steps:**
   1. Run `mkdir -p <TICKET_KEY>_attachments` in the repository root **first**, before calling any download tool. Verify it exists with `ls -d <TICKET_KEY>_attachments`.
   2. Call the Jira download tool **with its `target_dir` parameter set to the absolute path of `<TICKET_KEY>_attachments`**. Do not call it with only the issue key — omitting `target_dir` causes the tool to fall back to an unknown default location and return full file content inline instead, which for binary/zip files means the agent has to regenerate the entire base64 payload as output text before it can decode it. This is slow and pointless when the tool can write directly to disk itself. Always pass `target_dir` explicitly.
   3. After the call, run `ls -la <TICKET_KEY>_attachments/` and confirm the file count matches the number of attachments the ticket reported. If any are missing, that's a failure to fix now, not a gap to silently ignore. Do not fall back to manually decoding inline base64 content unless the tool genuinely has no `target_dir`-equivalent option — check its parameters before assuming that.
   4. For any `.zip` file in that folder: unzip it in place (`unzip <file>.zip -d <TICKET_KEY>_attachments/`), then `ls` again to confirm the extracted contents are there.
   5. Read/scan every file now sitting in the folder (logs, `.tf` files, screenshots, exported configs) before moving to Phase 2 — don't rely on tool-call output alone, the actual files on disk are the source of truth.
   6. **Phase 1 completion check:** before marking this phase done or moving to Phase 2, confirm you have actually run `ls -la <TICKET_KEY>_attachments/` in this session and seen real filenames in the output. If you haven't run that command yet, Phase 1 is not done — go back and run steps 1-4 for real, don't proceed on the strength of what the download tool's response showed you.

## Phase 2 — Analyze the Bug (No Code)

Do NOT read or touch any source code in this phase. Work purely from the ticket context, comments, and attachments.

1. **State the bug clearly**:
   - What is the **observed behavior** (what actually happens)?
   - What is the **expected behavior** (what should happen)?
   - What is the **impact** (who is affected, how badly)?

2. **Identify the trigger conditions**:
   - What specific resource/config triggers it?
   - What preconditions must exist (e.g., "ring must have members")?
   - Is it deterministic or intermittent?
   - Does it affect a specific provider version only?

3. **Identify the affected component**:
   - Which Terraform resource is involved (e.g., `genesyscloud_routing_queue`)?
   - What operation fails (create, update, delete, import, export)?
   - What API call is failing and with what error code?

4. **Form a hypothesis** about what's going wrong — based purely on the bug description, error messages, and attachment content. Explain this hypothesis to the user in plain language.

5. **Present the analysis** to the user and confirm understanding before proceeding.

## Phase 3 — Create Reproduction Config

After the analysis is confirmed, provide a Terraform configuration that reproduces the bug:

1. **Create a `.tf` file** that the user can use in their separate test folder (outside this repository). The user's test setup:
   - Has a folder where they run `terraform init`, `terraform plan`, `terraform apply`
   - Uses a local provider binary (built with `make build && make sideload` from this repository)
   - Provider source is `genesys.com/mypurecloud/genesyscloud` version `0.1.0`

2. **The config should**:
   - Mirror the exact scenario from the bug (same resource type, same attribute structure)
   - Use hardcoded IDs where needed (reference existing resources in the user's org) or create new ones if possible
   - Include comments explaining what each part does and why it's needed for reproduction
   - Include step-by-step instructions as comments (import existing resource if needed, what to change, what to expect)

3. **If reproduction requires multiple steps** (e.g., create first, then modify), provide them as separate configs or clearly labeled stages within one file.

4. **Tell the user**:
   - Which provider version/branch to build from (main branch first, then the specific tag if needed)
   - Any manual preconditions (e.g., "assign this skill to an agent in the UI")
   - What the expected error output looks like
   - How to verify the bug is present (e.g., `terraform plan -refresh=false` to check state corruption)

5. **STOP HERE. This is a hard blocking checkpoint, not a formality.** Do not investigate code, form conclusions about root cause, or write any fix until the user comes back with a result. Do not tell yourself the ticket evidence is "clear enough" to skip this — that reasoning is explicitly disallowed. The user will run the repro config themselves and report back with confirmation plus evidence (tfstate, `terraform plan`/`apply` output, error text). End your turn after presenting the repro config and instructions — do not continue into Phase 4 in the same turn or any later turn without a real user message confirming the result.

   - If they confirm reproduction: move to Phase 4, using their provided output as evidence for the hypothesis.
   - If they say it did *not* reproduce: don't guess at a new config silently — ask what actually happened (error vs no error vs different error) and revise the hypothesis from Phase 2 before generating a new repro attempt.

## Phase 4 — Investigate the Code

Only after the bug is reproduced (confirmed by the user, with evidence) — or confirmed unreproducible with explanation — proceed to investigate the source code:

1. Use the context-gatherer sub-agent to trace the relevant code paths.
2. Identify the root cause at a specific line/function level.
3. Present the root cause and proposed fix approach to the user before making changes.

## Phase 5 — Fix and Verify

1. Implement the fix.
2. Ensure `go build` and `go vet` pass.
3. Run relevant unit tests: `make testunit`
4. Have the user rebuild (`make build && make sideload`) and re-test with the reproduction config.
5. Summarize: root cause, what was fixed, what was verified.