# Milestone 2 retention and legal-hold schedule

These are engineering defaults pending written approval by the privacy owner/DPO under the Nigeria Data Protection Act 2023, NDPR 2019, and applicable NDPC guidance.

| Record | Default retention | Disposal |
| --- | --- | --- |
| Unconsumed email-verification token | 30 minutes plus at most 24 hours | Delete token row |
| Unconsumed password-recovery token | 15 minutes plus at most 24 hours | Delete token row |
| Consumed one-time token | At most 24 hours after consumption | Delete token row |
| Refresh session | 30-day absolute maximum; revoked records retained 30 days for replay investigation | Delete session row |
| Unverified account | 7 days | Begin approved account-erasure workflow |
| Security audit event | 12 months | Delete unless incident hold applies |
| Privacy-rights request receipt | 24 months after closure, subject to legal approval | Delete or irreversibly anonymise |
| Active profile | Account lifetime | Delete through approved account-erasure workflow |

Automated cleanup must run with a dedicated least-privilege database role, report counts rather than personal fields, and be idempotent. A documented legal hold suspends deletion only for the named record categories and expiry date; it does not authorize unrelated collection. Hold creation, extension, and release require an auditable privacy/legal owner decision.
