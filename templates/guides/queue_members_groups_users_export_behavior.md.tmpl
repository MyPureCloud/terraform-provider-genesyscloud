---
subcategory: ""
page_title: "Export Behavior for Queue Members – Avoid Reserved Ring Number Misconfiguration"
description: |
  Understand and avoid misconfigurations when assigning Users and Groups in Genesys Admin UI queues. Learn how reserved Bullseye Ring Numbers affect Terraform exports and how to correct issues.
---

# Export Behavior for Queue Members – Avoid Reserved Ring Number Misconfiguration

When configuring **Queue Members > Users or Groups** in the Genesys Admin UI, it's crucial to avoid assigning members to the **highest visible Bullseye Ring Number**. This number, while displayed in the UI, is **reserved** for internal system purposes and is **not included** in exported configurations.

## Problem Summary

In the Admin UI, you may see one more Bullseye Ring Number than you have manually configured. For example:

- If you configure two ring levels, the UI will show rings **1, 2, and 3**.
- Ring **3** is **automatically appended by the system** and is **reserved**.
- **Assigning users or groups to Ring 3 is invalid** and will not be captured in exports.

This misconfiguration may lead to:

- **Missing members in Terraform-exported configuration**
- **Confusion during CI/CD deployments**
- **Unexplained behavior in downstream systems**

## Why This Happens

This behavior is **by design**. The system internally reserves the highest ring number (N+1) for fallback logic or special routing conditions. It is **not meant for manual use**.

Terraform exporters **exclude** this ring from the `.tf` or `.tf.json` output to prevent invalid configurations.

## Example Scenario

| Configured Ring Numbers | Displayed in UI | Assigned to Users/Groups | Shown in Export |
|-------------------------|------------------|----------------------------|------------------|
| 1, 2                   | 1, 2, 3          | 1, 2 ✅, 3 ❌              | 1, 2 ✅          |

> ❌ Ring 3 is **not exported** and should not be used.

## Impact on Exports

If a user or group is accidentally assigned to the reserved ring:

- They **won’t appear** in Terraform-generated configuration.
- You’ll have a **false sense of completeness** in exported data.
- Future changes might **overwrite or exclude** those assignments.

## How to Correct This

1. **Audit queue member assignments** in the Admin UI.
2. **Reassign any users or groups** that are linked to the highest ring number (e.g., Ring 3 in the UI) to a lower, valid ring (e.g., 1 or 2).
3. Re-run the Terraform export to verify inclusion.

## Best Practices

- Always configure ring levels explicitly.
- Only assign Users/Groups to those ring levels.
- **Ignore the highest ring shown** in the UI—it is **not a usable value**.

## Summary

| Do | Don't |
|----|-------|
| Assign users/groups to configured ring numbers only | Use the last visible ring number in Admin UI |
| Audit UI vs. exported configuration regularly | Assume all UI-visible values are valid |
| Follow Terraform export validations and logs | Expect reserved ring numbers to export |

By understanding and avoiding this common misconfiguration, you ensure accurate, predictable, and reliable queue member exports.

---
