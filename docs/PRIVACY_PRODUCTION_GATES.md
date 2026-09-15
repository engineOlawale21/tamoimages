# Privacy production gates

Engineering controls in this repository do not authorize production processing. The accountable organisation must complete and retain the following approvals in its restricted governance system.

## Ownership and regulatory assessment

- Name the data-processing owner, security owner, incident commander, data-subject-request owner, and DPO or privacy counsel.
- Record the controller/processor role separately for account management, contributor uploads, model/property releases, catalogue publication, licensing, payments, support, analytics, and observability.
- Decide and document whether the organisation is a Data Controller or Processor of Major Importance, whether registration is required, and which compliance-audit filings apply.
- Approve every processing-inventory lawful basis and retention period.

## Vendors and transfers

Maintain a restricted subprocessor register with vendor, service, data categories, subjects, purpose, role, recipients, storage and support countries, transfer basis, safeguards, agreement date, review date, deletion commitment, incident deadline, audit rights, confidentiality terms, and downstream subprocessors. No production personal data may be sent until security/privacy due diligence and the processing agreement are approved.

## DPIA gate

A DPO-approved DPIA is required before processing sensitive or child data, biometrics, precise location, large-scale monitoring or profiling, significant automated decisions, or AI analysis of faces, identity, or private media. Record necessity, proportionality, data flows, risks to people, mitigations, residual risk, consultation, owner, approval, and review date.

## Operations evidence

- Configure TLS and encryption at rest using the selected production platform.
- Use managed secrets, documented rotation, and least-privilege service identities.
- Schedule retention jobs and alert on failures; retain tamper-resistant deletion completion records.
- Approve audit-log access, integrity protection, retention, dashboards, and alerts.
- Review dependency and container scan results, remediate findings, and retain evidence.
- Test backup restoration and the personal-data incident workflow before launch and at least annually.
- Monitor ClamAV definition freshness and scan-failure rates; prevent publication unless the scan succeeds.
- Publish approved privacy contact details and replace the engineering privacy-notice wording before production.
