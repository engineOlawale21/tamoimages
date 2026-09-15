# Media-service data-processing inventory

This is an engineering inventory. The privacy owner/DPO must approve lawful bases, retention periods, publication rules, processor locations, and cross-border safeguards before production.

| Data | Purpose | Owner | Location | Retention class | Deletion path |
| --- | --- | --- | --- | --- | --- |
| Contributor ID | Associate media with its owner without copying identity profile data | Media and privacy owners | Media PostgreSQL | Asset/account lifecycle | Soft deletion followed by approved hard-deletion workflow |
| Original filename | Upload administration and contributor display | Media owner | Media PostgreSQL | Private upload metadata | Delete or replace during asset deletion/anonymisation |
| Storage key | Locate the private source object | Media owner | Media PostgreSQL | Asset lifecycle | Delete database locator after confirmed object deletion |
| Content type and byte size | Validation, processing, and delivery decisions | Media owner | Media PostgreSQL | Asset lifecycle | Delete with asset metadata |
| Uploaded media object | Processing, review, publication, and licensing | Media, licensing, and privacy owners | S3-compatible private object storage | Status- and license-specific schedule | Object deletion workflow propagated to generated variants and processors |
| Extracted technical metadata | Validate and present media dimensions, duration, and codec without duplicating content | Media owner | Media PostgreSQL | Asset lifecycle | Delete with asset metadata |
| Generated image variants | Responsive contributor review and future catalogue delivery | Media owner | S3-compatible private object storage | Derived-asset lifecycle | Delete before or with the source asset; clean partial processing output on failure |
| Generated video renditions and poster frames | Efficient preview, review, and future catalogue playback | Media owner | S3-compatible private object storage | Derived-asset lifecycle | Delete before or with the source asset; clean partial processing output on failure |
| Model and property release forms (maximum three per media asset) | Evidence of permission for review, publication, and licensing | Media and privacy owners | S3-compatible private object storage; minimal status metadata in PostgreSQL | Media lifecycle and documented legal/licensing retention | Private access only; expired incomplete uploads do not retain an active slot; delete with the governed source lifecycle unless legal hold applies |
| Batch review feedback and decisions | Contributor submission review and correction workflow | Media owner | Media-service PostgreSQL | Batch and media lifecycle | Contributor-visible feedback is bounded to 2,000 characters and must not contain private reviewer notes |

Original filenames, storage keys, private object URLs, and media contents are prohibited in operational logs. Public release and licensed-record retention must be modeled separately from temporary, rejected, failed, and deleted uploads.

Contributor deletion removes the original, generated variants, private release documents, and batch links before scrubbing personal metadata. The service retains only append-only completion evidence containing opaque identifiers, reason, object count, completion time, and a SHA-256 digest of the sorted deletion targets; raw storage keys are not retained in that evidence.

The processing worker invokes ClamAV before metadata extraction or derivative generation and fails closed on an infected file or unavailable scanner. Production image builds refresh malware definitions; operations must also update definitions continuously and alert on stale definitions or scan failures.
