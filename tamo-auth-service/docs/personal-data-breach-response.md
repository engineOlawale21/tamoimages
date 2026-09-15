# Personal-data breach response ownership

This runbook supports the repository's NDPR security standard. It is an engineering control and does not replace advice from the organisation's Data Protection Officer (DPO) or legal counsel.

## Ownership

- The on-call backend engineer contains the incident, preserves evidence, and immediately informs the incident commander.
- The incident commander coordinates engineering, security, operations, communications, and executive stakeholders.
- The DPO determines whether personal data is involved, records the decision, assesses risk to data subjects, and owns regulatory and data-subject notification decisions.
- Service owners identify affected systems, processing purposes, data categories, subjects, processors, and retention locations.
- Only the designated communications or legal owner contacts regulators, customers, contributors, or the public.

Named individuals and current contact channels belong in the private incident-management system, not this public repository.

## Required response

1. Open a restricted incident record with time discovered, reporter, correlation IDs, affected services, and known exposure window.
2. Contain access without destroying logs or other evidence. Rotate exposed credentials and invalidate affected sessions when appropriate.
3. Preserve an append-only decision timeline. Do not copy personal data into tickets, chat, or logs.
4. Determine the affected data, subjects, processing purposes, geographic scope, processors, and likely consequences.
5. The DPO records the notification assessment and applicable deadline. Escalate immediately; do not wait for full root-cause certainty.
6. Notify approved parties through authorised channels when the DPO or legal owner directs it.
7. Eradicate the cause, restore safely, increase monitoring, and validate access controls.
8. Complete root-cause analysis, corrective actions, evidence retention, and the data-processing inventory update.

## Minimum audit record

The restricted record must contain who made each decision, when it was made, evidence considered, systems and data affected, containment actions, notification reasoning, recovery validation, and follow-up owners. Secrets, raw tokens, passwords, and unnecessary personal data must never be included.
