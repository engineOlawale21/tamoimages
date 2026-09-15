# Backend Security and Nigerian Data-Protection Standard

## Status and authority

This standard is mandatory for `tamo-auth-service` and `tamo-media-service`. It implements engineering safeguards for the Nigeria Data Protection Act 2023, the Nigeria Data Protection Regulation 2019 (NDPR), and applicable current guidance issued by the Nigeria Data Protection Commission (NDPC).

The NDP Act 2023 and binding current NDPC instruments take precedence if an older NDPR provision conflicts with them. The DPO or designated Nigerian privacy counsel must approve the production processing model, notices, retention schedule, processor arrangements, cross-border transfer mechanisms, incident decisions, and regulatory filings.

Primary references:

- [Nigeria Data Protection Act 2023](https://ndpc.gov.ng/wp-content/uploads/2024/03/Nigeria_Data_Protection_Act_2023.pdf)
- [NDPC resources, including NDPR 2019](https://www.ndpc.gov.ng/resources/)
- [NDPC compliance FAQs](https://www.ndpc.gov.ng/faqs/)

## Required controls

### Accountability and processing inventory

- Assign a named owner to every processing activity.
- Maintain a record of processing containing data categories, data subjects, purpose, lawful basis, recipients/processors, locations, retention, safeguards, and deletion method.
- Identify controller and processor roles per workflow instead of applying one role to the whole platform.
- Assess Data Controller or Processor of Major Importance registration, DPO, and compliance-audit obligations before production.
- Store policy and notice versions so the platform can prove what applied at a given time.

### Lawfulness, transparency, and consent

- No endpoint may collect a personal-data field without a documented purpose, lawful basis, and retention class.
- Privacy notices must be concise, accessible, versioned, and presented at or before collection.
- Consent must be granular and unbundled where relied upon. Record subject, notice version, purpose, choice, timestamp, and collection channel.
- Reject dark patterns. Withdrawal must be available through the same account channel and must stop future consent-based processing.

### Minimisation and purpose limitation

- DTOs use allowlists and reject unknown properties.
- Do not copy identity data into the media database when an opaque user identifier is sufficient.
- Events contain the minimum attributes required by consumers; consumers may not create shadow user profiles.
- Development fixtures and telemetry must not contain production personal data.

### Data-subject rights

- Provide verified request flows for access, correction, deletion, restriction, objection, portability, consent withdrawal, and complaints.
- Prevent account takeover during rights requests with proportionate identity verification.
- Track receipt, verification, decision, fulfilment, deadline, exemptions, and reviewer without logging request contents unnecessarily.
- Propagate approved correction/deletion actions to databases, object storage, caches, search indexes, processors, and eligible backups.

### Retention and deletion

- The data owner must approve a retention schedule before a table, object prefix, topic, or durable log containing personal data is created.
- Enforce expiry with scheduled jobs and produce deletion metrics and immutable completion evidence.
- Use irreversible anonymisation only when re-identification is not reasonably possible.
- Legal holds must identify scope, authority, approver, start, review date, and release date.

### Security safeguards

- TLS is mandatory outside isolated local development; encrypt databases, object storage, and backups at rest.
- Use least-privilege service identities and separate identity/media database ownership.
- Store production secrets in a managed secret store and rotate them; never commit secrets.
- Hash passwords with an approved adaptive password hash and unique salts.
- Use short-lived access tokens, rotating refresh sessions, revocation, issuer/audience validation, and reuse detection.
- Validate content signatures, size, and type; malware-scan uploads before publication.
- Patch dependencies and base images, scan artifacts, test backups, and exercise restoration.

### Logging and audit

- JSON logs include timestamp, severity, service, environment, event name, request/correlation ID, route template, status, duration, and pseudonymous actor ID where needed.
- Logs must exclude authorization headers, cookies, passwords, tokens, reset/verification codes, payment details, government identifiers, private object URLs, and request/response bodies by default.
- Security audit events record actor, action, target identifier, result, timestamp, correlation ID, and source context without recording secret values.
- Restrict and monitor audit-log access; define a retention period and integrity control.

### Processors and cross-border transfers

- Complete security/privacy due diligence and a processing agreement before sending personal data to a vendor.
- Maintain a subprocessor register covering hosting, storage, email, analytics, payments, support, and observability.
- Record processing countries and obtain DPO/legal approval for the applicable cross-border transfer basis and safeguards before transfer.
- Vendor deletion, incident reporting, audit, confidentiality, and subprocessor duties must be contractually addressed.

### DPIA and privacy review

A documented DPIA and DPO/privacy approval are release gates for likely high-risk processing, including sensitive data, child data, biometrics, large-scale monitoring/profiling, automated decisions with significant effects, precise location, or AI analysis of faces, identity, or private media.

### Personal-data incidents

- Maintain an incident channel, on-call owner, severity model, containment playbook, evidence log, and processor contact list.
- Record discovery time, affected systems/data/subjects, likely consequences, containment, risk decision, notification decision, and approvals.
- The DPO/legal owner decides regulatory and data-subject notification requirements and timing under current law; engineering must make prompt detection, investigation, export, and notification possible.
- Exercise the workflow at least annually and remediate findings.

## Milestone 1 release gates

- Configuration fails closed and production secrets meet minimum strength requirements.
- Input allowlisting, size limits, CORS allowlisting, security headers, rate-limit foundations, and generic production errors are enabled.
- Correlation IDs are returned and propagated without exposing credentials or sensitive bodies.
- Liveness and dependency readiness are separate.
- Service containers run without root privileges where possible.
- CI runs tests, lint/static analysis, builds, OpenAPI validation, and image builds.
- Each new personal-data field has a purpose, lawful basis owner, retention class, and deletion path recorded.
- A breach-response owner and data-subject-request owner are named before production deployment.
