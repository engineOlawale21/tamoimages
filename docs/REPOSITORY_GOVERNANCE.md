# Repository governance and releases

## Ownership

`CODEOWNERS` supplies default path ownership. Replace the placeholder team with the organisation's actual GitHub team before enabling mandatory review. Security and NDPR changes also require review by the designated security owner or DPO delegate through the private review process.

## Branch protection

Protect `main` with pull requests, at least one approving review, dismissal of stale approvals, resolved conversations, linear history, and required current CI checks. Administrators should not bypass controls except through the documented emergency process. Force pushes and branch deletion are disabled.

## Versions and releases

The repository uses Semantic Versioning. Each deployable has an independent release version and image tag even though source shares one Git history. Tags use:

- `auth-vMAJOR.MINOR.PATCH`
- `media-vMAJOR.MINOR.PATCH`
- `web-vMAJOR.MINOR.PATCH`
- `infrastructure-vMAJOR.MINOR.PATCH`

A breaking external API or event-schema change increments the owning deployable's major version. Backward-compatible functionality increments minor; fixes increment patch. Shared contract changes identify every affected deployable in the pull request and changelog.

## Changelog

User-visible, operational, security, schema, API, event, and migration changes are recorded under the relevant deployable in root `CHANGELOG.md`. Entries use `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, and `Security` headings. Pull requests update `Unreleased`; a release moves entries into a dated version section.

## Initial baseline

The baseline commit is created only after secrets are scanned and the clean-install verification has passed. Generated dependencies, build output, local environments, and personal IDE state are excluded.

## Dependency updates

Dependency updates are made through a pull request with regenerated lockfiles, release-note and advisory review, unit/integration tests, production builds, and container scanning. Major upgrades require an explicit migration note and rollback plan. Emergency security updates may use the expedited review path but must still preserve lockfiles and verification evidence. Automated update tools receive read-only defaults and may not merge or publish releases without the required reviews.
