# Roles & Permissions

Users gain site access via `sl-cli user grant <site> <email> [--role <role>]`.

| Action | admin | editor | viewer |
|--------|-------|--------|--------|
| View admin editor | ✅ | ✅ | ✅ |
| Save content (pages, blog, forms) | ✅ | ✅ | ❌ |
| Create/publish blog posts | ✅ | ✅ | ❌ |
| Delete blog posts / files | ✅ | ❌ | ❌ |
| View form submissions | ✅ | ✅ | ✅ |
| Delete form submissions | ✅ | ❌ | ❌ |
| Manage users (grant/revoke) | ✅ | ❌ | ❌ |
| Admin tokens (create/revoke) | ✅ | ❌ | ❌ |

**RBAC enforcement:**
- **UI layer:** Save/Publish/Delete buttons hidden or replaced with "Read-only" spans based on role
- **Runtime:** `savePost()`, `deletePost()`, `saveForm()`, `deleteSubmission()` reject with toast if role insufficient
- **Role resolved:** From `sl_admin_session` JWT → `site_users` table at editor load time

## Commands

```bash
# Grant access (default role: admin)
sl-cli user grant <site> <email> --role admin|editor|viewer

# Revoke all access
sl-cli user revoke <site> <email>

# Dashboard shows only granted sites with role badge
# GET /admin → login → list of sites
```
