# Identity-service data-processing inventory

This engineering inventory must be reviewed and approved by the privacy owner/DPO before production. Lawful-basis decisions are intentionally assigned to that owner rather than inferred by application code.

| Data | Purpose | Proposed lawful-basis owner | Location | Retention class | Deletion path |
| --- | --- | --- | --- | --- | --- |
| Email address | Authentication, essential account communication, security recovery | Privacy/legal owner | Identity PostgreSQL only | Active account | Soft-delete, then scheduled hard deletion after the approved recovery/legal-hold window |
| First and last name | Account identification and attribution | Product and privacy owners | Identity PostgreSQL only | Active account | Same account-deletion workflow |
| Account role | Authorization and buyer/contributor workflow selection | Product owner | Identity PostgreSQL; minimum role claim in access token | Active account | Same account-deletion workflow; access token expires independently |
| Password hash | Authenticate the account | Security and privacy owners | Identity PostgreSQL only | Active account | Delete with account or immediately replace on password reset |
| Correlation ID | Diagnose requests and investigate security events | Security owner | Restricted operational logs | Short operational-log class | Log lifecycle expiration |
| Pseudonymous network-rate key | Prevent credential attacks without retaining raw IP addresses | Security and privacy owners | Redis only; HMAC-derived value | Authentication rate-limit window | Automatic Redis expiry |
| Verification token hash | Prove email ownership | Security and privacy owners | Identity PostgreSQL | 30-minute validity plus no more than 24 hours | Scheduled token-row deletion |
| Recovery token hash | Authorize password replacement | Security and privacy owners | Identity PostgreSQL | 15-minute validity plus no more than 24 hours | Scheduled token-row deletion |
| Refresh-session token hash and family | Maintain sessions and detect replay | Security owner | Identity PostgreSQL | 30-day absolute maximum; revoked evidence 30 days | Scheduled session-row deletion |
| Hashed user-agent and network prefix | Investigate session misuse without invasive fingerprinting | Security and privacy owners | Identity PostgreSQL | Same as refresh session | Delete with session |
| Security event type, outcome, subject hash | Detect and investigate authentication abuse | Security and privacy owners | Identity PostgreSQL | 12 months unless documented legal hold | Scheduled event deletion |
| Role-specific profile | Provide buyer or contributor features | Product and privacy owners | Identity PostgreSQL | Account lifetime | Account-erasure workflow |
| Privacy request receipt | Demonstrate and operate data-subject rights | Privacy/legal owner | Identity PostgreSQL | 24 months after closure, pending approval | Scheduled deletion/anonymisation |
| Privacy notice version and optional consent choice | Prove transparency and the subject's current granular choice | Privacy/legal owner | Identity PostgreSQL | Notice versions: permanent policy evidence; consent evidence: account lifetime plus approved claims window | Append withdrawals; delete or irreversibly anonymise subject linkage after the approved claims window |

Prohibited in operational logs: email, names, passwords, password hashes, authorization headers, tokens, cookies, verification/reset codes, request bodies, and database connection strings.

Before adding any field, update this inventory with its purpose, lawful-basis owner, recipients, location, retention, and deletion propagation requirements.
