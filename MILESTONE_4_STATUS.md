# Milestone 4 status

Last verified: 8 September 2026.

## Completed and verified

- Contributor-owned media libraries persist processing status and signed variants.
- Ready media supports validated title, description, keywords, location, and creative/editorial metadata.
- Contributors can create draft batches, assign owned ready media once, and submit metadata-complete batches.
- A media asset supports up to three private model/property release forms; expired pending uploads stop consuming a slot.
- Batch submission publishes `media.submitted-for-review.v1` and review decisions publish `media.review-completed.v1`.
- Reviewer transitions are limited to submitted/in-review batches; rejection requires bounded contributor-visible feedback.
- The contributor dashboard exposes batch creation, assignment, submission, release uploads, statuses, failures, retries, and review feedback.
- The live PostgreSQL acceptance journey passes: processed media -> metadata -> three releases -> fourth release rejected -> batch assignment -> submission -> rejection feedback.

## Remaining release gates

- Browser automation for the complete contributor journey, including direct MinIO upload and keyboard/screen-reader checks.
- A production reviewer-account provisioning mechanism; public registration remains limited to buyers and contributors.
- Transactional outbox/inbox delivery and dead-letter handling, scheduled for the production-reliability milestone.
- Privacy/legal approval of release-document retention and reviewer operating procedures.

Milestone 4 engineering functionality is complete enough to begin Milestone 5 development. The gates above must pass before production release.
