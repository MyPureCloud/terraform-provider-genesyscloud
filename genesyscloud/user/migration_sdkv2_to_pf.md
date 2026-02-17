**User Perspective (SDKv2 → Plugin Framework)**

> **Purpose of this file**
> This document captures **user-visible behavior changes** caused by migrating this resource from **Terraform SDK v2** to the **Terraform Plugin Framework (PF)**.
>
> ⚠️ **Important**
>
> * Do **not** include provider implementation details
> * Do **not** include Go code, schema internals, or framework mechanics
> * Focus strictly on **Terraform user experience and behavior**
> * Follow the **KISS principle**

---

## 1. What Worked in SDKv2

* Removing the `addresses` block from configuration appeared to work without errors
* `terraform apply` completed successfully when addresses block was omitted
* Phone numbers were deleted from user addresses when addresses block was removed
* Test cases for address removal (DEVTOOLING-1238) passed consistently

**Example Configuration:**

```hcl
# Step 1: Create user with addresses
resource "genesyscloud_user" "example" {
  name  = "test_user"
  email = "test@example.com"
  
  addresses {
    phone_numbers {
      media_type = "PHONE"
      number     = "+13175559000"
      type       = "WORK"
    }
    other_emails {
      address = "work@example.com"
      type    = "WORK"
    }
  }
}

# Step 2: Remove addresses block
resource "genesyscloud_user" "example" {
  name  = "test_user"
  email = "test@example.com"
  # addresses block removed - apply succeeds
}
```

**Resulting State (terraform.tfstate):**

```json
{
  "addresses": [
    {
      "other_emails": [
        {
          "address": "work@example.com",
          "type": "WORK"
        }
      ],
      "phone_numbers": []
    }
  ]
}
```

**Observation:** Phone numbers deleted, but EMAIL contacts remained in state. No errors reported.

---

## 2. What Does Not Work in Plugin Framework (PF)

* Removing the `addresses` block from configuration fails with consistency error when EMAIL contacts exist
* `terraform apply` fails with "Provider produced inconsistent result after apply" error
* Test cases for complete address removal (DEVTOOLING-1238) fail at address deletion steps

**Same Configuration as SDKv2 Example:**

```hcl
# Step 1: Create user with addresses
resource "genesyscloud_user" "example" {
  name  = "test_user"
  email = "test@example.com"
  
  addresses {
    phone_numbers {
      media_type = "PHONE"
      number     = "+13175559000"
      type       = "WORK"
    }
    other_emails {
      address = "work@example.com"
      type    = "WORK"
    }
  }
}

# Step 2: Remove addresses block
resource "genesyscloud_user" "example" {
  name  = "test_user"
  email = "test@example.com"
  # addresses block removed - apply FAILS
}
```

**Error Message:**

```
Error: Provider produced inconsistent result after apply

When applying changes to genesyscloud_user.example, provider 
produced an unexpected new value: .addresses: block count 
changed from 0 to 1.

This is a bug in the provider. Please report it.
```

**What Happened:** Phone numbers were deleted, but EMAIL contacts remained. Plugin Framework detected the mismatch between expected state (no addresses) and actual state (EMAIL contacts present).

---

## 3. Behavioral Changes

* Plugin Framework detects state inconsistencies that SDKv2 did not report
* EMAIL contacts remaining in state after address removal now causes apply failures instead of silent acceptance
* Error messages are more explicit about configuration vs actual state mismatches

**Comparison:**

| Scenario | SDKv2 Behavior | Plugin Framework Behavior |
|----------|----------------|---------------------------|
| Remove addresses block with phone numbers only | ✅ Success - all deleted | ✅ Success - all deleted |
| Remove addresses block with EMAIL contacts | ✅ Success - but EMAIL remains in state | ❌ Fails - consistency error |
| User expectation | All addresses deleted | All addresses deleted |
| Actual result | Phone deleted, EMAIL remains | Apply fails before state update |

---

## 4. Limitations

* Complete address deletion by omitting the `addresses` block is not currently supported when EMAIL contacts exist
* Stricter state consistency validation may expose underlying API behavior that was previously hidden

---

## 5. Known Gaps

* Address deletion behavior differs from SDKv2 when EMAIL contacts are present
* Workaround for complete address removal may be required in some scenarios

---

## 6. Impact on Existing Configurations

* Configurations that remove the `addresses` block may fail after upgrading if EMAIL contacts exist
* Users should review address management workflows before migration
* Terraform apply operations involving address removal should be tested in non-production environments first
* State files may show EMAIL contacts remaining after address block removal attempts

**Migration Scenario Example:**

**Before Migration (SDKv2):**
```hcl
# This configuration works in SDKv2
resource "genesyscloud_user" "existing_user" {
  name  = "existing_user"
  email = "user@example.com"
  # addresses block removed to "delete" all addresses
}
```

**After Migration (Plugin Framework):**

If the user had EMAIL contacts before addresses block removal:
```
❌ terraform apply fails with consistency error
```

**Recommended Approach:**

Check your state file before migration:
```bash
terraform show | grep -A 10 "addresses"
```

If EMAIL contacts exist in state, update configuration to explicitly manage them or use alternative deletion approach.

---

## 7. Current Status

Provide a calm closing statement.

**Template text (do not over-edit):**

> This resource has been migrated to the Terraform Plugin Framework. Core functionality is supported, and the listed differences are known. Improvements will be delivered incrementally in future releases.

---
