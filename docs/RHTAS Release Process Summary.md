# RHTAS Release Process Summary

Summary of the [RHTAS Release Process Draft](RHTAS%20Release%20Process%20Draft.md).

---

## Build Dependency Chain

Bottom-up build order:

1. **Layer 1 -- Base Components**
   - fulcio, rekor, trillian, cosign (CLI tools), gitsign, timestamp-authority, certificate-transparency-go, rekor-monitor, rekor-search-ui, tough, segment-backup-job
   - Build independently in Konflux
   - Successful builds trigger "nudges" (automated PRs) to downstream repos to update image references

2. **Client-Server Bundle**
   - Lives in the cosign repo
   - Aggregates CLI tools (cosign, gitsign, rekor-cli, etc.) into a single Konflux bundle
   - No separate layer 2 repo

3. **Layer 3 -- Main Products**
   - `secure-sign-operator` and `artifact-signer-ansible`
   - Consume the client-server bundle and direct layer 1 images

4. **FBC (File-Based Catalog)**
   - Built via GitHub Action, not Konflux pipeline
   - Per-OCP-version catalogs
   - OLM distribution mechanism for the operator

---

## Release Candidate Curation

Happens in the [`releases`](https://github.com/securesign/releases) repo:

1. **Promote-to-candidate pipeline** collects successful Konflux snapshots into candidate JSON files on a bot branch (`RHTAS-build-bot_candidate-images`):
   - `component-*.json` -- layer 1 component images
   - `operator-*.json` -- operator image
   - `fbc-*.json` -- FBC image

2. **Snapshot validation** -- `validate-latest-snapshot` task ensures only the latest snapshot proceeds

3. **Push to GitHub** -- candidate JSON is committed to the releases repo

4. **Structural tests** validate the release candidate

5. **QE handover** for verification

---

## Release Plan Admissions (RPAs)

- Stored on internal GitLab (`gitlab.cee.redhat.com`)
- Map quay.io images to registry.redhat.io counterparts
- Define version tags applied to released images
- Must be updated for each release cycle

---

## Running the Release

1. `oc apply` ReleasePlans to the Konflux cluster
2. `oc apply` Release objects in correct order
3. Release types:
   - **RHEA** -- general release (errata advisory)
   - **RHSA** -- security advisory (includes CVE fixes)

---

## PCO (Policy Controller Operator)

- Separate from the main TAS release process
- Builds on Konflux but post-build steps are manual:
  - FBC generation
  - Git tagging
  - Release coordination

---

## MVO (Model Validation Operator)

Three Konflux components with automated build chain on push to `main`:
- `model-validation-operator` (Operator)
- `model-validation-operator-agent` (Agent)
- `model-validation-operator-bundle` (Bundle)

Build order: Agent -> Operator -> Bundle (automated via nudging)

Post-build steps are manual:
- FBC generation for OCP versions 4.16 through 4.21
- PR to `releases` repo with FBC + MVO component snapshots
- Submitting the PR triggers a staging run
- RPA on GitLab must be manually updated

---

## Comet

Container management system where RHTAS images are listed. Rarely needs changes since most images are already registered.

For new image onboarding:
- Type: Layered, Operator, or Operator bundle
- Build Source: Konflux Request
- Fill in Metadata, Configuration, Contacts
