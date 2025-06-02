# Release process

This document outlines the process for releasing a new version of the Mezo chain client.

## Determine the release type and version number

The `mezod` repository follows the semver versioning scheme. The first step
of every release is to determine its type and the new version number.

Compare the Git diff between the planned version (recent main) and the current version running on
the upgraded chain. Use the following rules to determine the version number:

- If the diff contains consensus breaking changes, the release is a major version upgrade.
- If the diff does not contain consensus breaking changes, the release is a minor/patch version upgrade.

Next, determine the optional suffix:

- If the release is a stable release ready to be rolled out on the Mainnet chain, the version number
  should not contain any suffix.
- If the release is a candidate release intended to be rolled out on the Testnet chain only,
  the version number should be suffixed with `-rcN`, where `N` is the release candidate number.

Ideally, new releases should be rolled out as candidate releases on the Testnet chain first.
Once they are tested and promoted to stable releases, they can be rolled out on the Mainnet chain.
In justified cases, the candidate release phase and even the Testnet rollout can be skipped, and
the given release may be rolled out on Mainnet directly, as a stable release.
Always use your best judgement when deciding about the release strategy.

## Prepare the release

### Major releases with consensus breaking changes

Major releases with consensus breaking changes require special preparations and adherence to the
on-chain upgrade process described in the [upgrades.md](./upgrades.md) document.
The steps are outlined below:

#### 1. Choose the on-chain upgrade type based on the [upgrades.md](./upgrades.md) document

In most cases, the **Planned upgrade with chain halt** type will be the right option.
This is the safest way though it requires a coordinated chain halt.

#### 2. Create the upgrade handler

Open a pull request named "The vX.Y.Z upgrade handler" in `mezod`. This PR must:

- Introduce the vX.Y.Z upgrade handler according to the chosen upgrade type. The handler
  must perform all necessary store migrations and state changes required by the upgrade.
  However, if the upgrade affects only the consensus logic, this handler can be a no-op.
- Update the historical upgrades table in [upgrades.md](./upgrades.md#historical-upgrades).
- Provide test instructions allowing to simulate the upgrade on localnet.
- Request at least two reviewers to simulate the upgrade on localnet.

Example: [The v2.0.0 upgrade handler](https://github.com/mezo-org/mezod/pull/492)

#### 3. Execute pre-release smoke tests

> [!NOTE]
> This step should be fully automated in the future.

Before cutting the tag, it's important to ensure that the binary and Docker image can be built
without errors. The binary part is handled by the `build.yml` workflow executed by the CI
system upon each PR. The Docker part must be tested manually by running `make build-docker-linux`
locally.

This step should prevent failures of the automatic processes that build the binary and Docker image
upon a new tag. A failure at this stage would be problematic as it would require overwriting the
tag or creating a new one.

#### 4. Cut and publish the new `mezod` tag

Cut and publish the new vX.Y.Z tag in the `mezod` repository.

Confirm that the following happened automatically:

- Docker image was built and published as expected
- Binary (amd64) was built and published as expected
- Draft release was created on GitHub (this happens only for stable releases)

Note that stable and candidate releases have different target locations for the binary and Docker image,
as pointed in [this document](https://github.com/mezo-org/validator-kit?tab=readme-ov-file#artifacts).

Creating the `release/vX.Y.Z` branch is optional. This can be always done later if there is a need to
backport newer changes to this release line.

#### 5. Execute post-release smoke tests

> [!NOTE]
> This step should be fully automated in the future.

Download the published binary and use it to run a single-node localnet to confirm everything works
as expected. Note that the binary is built for linux/amd64 only, so you may need a VM to execute
this test (there is one playground VM available in the `mezo-staging` GCP project).

Testing steps:

1. Get the vX.Y.Z binary from the appropriate location.
2. Modify the [`localnode-start.sh`](../scripts/localnode-start.sh) script to use the new binary:
   - Remove the `make install` call.
   - Replace all `mezod` invocations with the new binary path.
3. Run the `localnode-start.sh` script.
4. Confirm the node works and produces blocks.

#### 6. Publish the release on GitHub (only for stable releases)

Stable releases must be published on GitHub. At this point, a draft release is already here
as it was created by the CI system upon pushing the stable version tag.

Fill up the draft release according to the established pattern. Populate the changelog section
with GitHub-generated release notes (this requires setting the previous tag correctly).

Example: [v2.0.0 release](https://github.com/mezo-org/mezod/releases/tag/v2.0.0)

#### 7. Update the Validator Kit (only for stable releases)

Open a pull request named "Bump mezod to vX.Y.Z" in `validator-kit`. This PR must:

- Increment the version of the Helm chart (`version` field in [`Chart.yaml`](https://github.com/mezo-org/validator-kit/blob/main/helm-chart/mezod/Chart.yaml)).
  The chart version should change in the same way as the new `mezod` version, i.e. if `mezod` was increased
  to a new major, the chart version should be increased to a new major as well.
- Update `mezod` app version in the Helm chart to vX.Y.Z (`appVersion` field in `Chart.yaml`)
- Update `mezod` default image tag in the Helm chart to vX.Y.Z
  (`tag` field in [`values.yaml`](https://github.com/mezo-org/validator-kit/blob/cc88601bfbef41c844a9e81d79db9e53721e3761/helm-chart/mezod/values.yaml#L2))
- Update the version ordering in the [Node synchronization](https://github.com/mezo-org/validator-kit?tab=readme-ov-file#node-synchronization)
  section of the root `README.md`

Once the PR is merged, cut and publish a new tag in the `validator-kit` repository.
The tag should be the same as the updated Helm chart version (`version` field in `Chart.yaml`).

Example: [Bump mezod to v2.0.0](https://github.com/mezo-org/validator-kit/pull/74)

### Minor/patch releases without consensus breaking changes

For minor/patch releases, the release process is basically the same as for major releases.
The only difference is that it does not require the on-chain upgrade process.
Changes can be rolled out immediately.

## Communicate the release

### Governance

Major releases with consensus breaking changes rolled out using the **Planned upgrade with chain halt**
upgrade type require that the chain governance issues an upgrade plan to the [`Upgrade` precompile](./upgrades.md#the-upgrade-precompile).

For Mainnet, use [safe.mezo.org](https://safe.mezo.org/home?safe=mezo:0x98D8899c3030741925BE630C710A98B57F397C7a)
to craft a `submitPlan` transaction JSON from the Mezo Governance SAFE, to the `Upgrade` precompile.

Fill transaction fields as follows:

- The `name` must be vX.Y.Z. This must exactly match the name of the upgrade handler
- The `height` must be the block at which the chain should halt (use BlockScout to see estimated date of future blocks)
- The `info` should point to the binary download link in the Cosmovisor format. For example:
  `{"binaries":{"linux/amd64":"https://github.com/mezo-org/mezod/releases/download/v2.0.0/linux-amd64.tar.gz"}}`

Pass the JSON to the Mezo Governance SAFE signers and make sure the transaction is executed BEFORE the
planned chain halt block.

For Testnet, this process is simpler as the governance is an EOA.

### Node operators

Validators and other node operators should be notified about the new release and the upcoming
on-chain upgrade (if applicable).

Currently, this communication is issued through the [validator-alerts](https://discord.com/channels/1220035427952627863/1319326473991098440)
Discord channel.

The message depends on the release type and on-chain upgrade mechanism used. In general, the message should:

- Have a visible title clearly denoting that this is a release announcement
- Mention the new version and link to the GitHub release notes containing a detailed changelog
- Inform about the impacted chain (Mainnet, Testnet, or both)
- Clearly state what actions should be taken by node operators and when
- Denote which types of nodes are affected (any combination of: validator, RPC, seed)
- Link to the pre-built binary and Docker image artifacts
  (i.e., [Artifacts section in the Validator Kit README](https://github.com/mezo-org/validator-kit?tab=readme-ov-file#artifacts))

Example: [v2.0.0 announcement](https://discord.com/channels/1220035427952627863/1319326473991098440/1377657996640915656)

### Community

If the release impacts end users, a community-facing announcement should be made.
This is typically handled by the marketing team. Engineering team is in charge of providing
the necessary information.

## Resources

### Linear templates

- [Mezod Release Checklist Template](https://linear.app/thesis-co/team/TET/new?template=02a8c718-b654-406d-bc08-06de1d7524f3)
