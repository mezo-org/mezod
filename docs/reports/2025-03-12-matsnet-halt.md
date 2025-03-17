# Incident report: 2025-03-12 Matsnet halt

In the afternoon UTC of March 12, 2025, a transaction was submitted to the
Mezo testnet, referred to as Matsnet, that caused a consensus failure in the network
and a halt to block production. The system was restored to working
status roughly 22 hours after the incident began by an already-planned upgrade
whose timeline was accelerated.

The incident response time was deliberately slower than a standard production
response would have been given the testing nature of the network, and the team
used the opportunity to test our incident response practices and identify gaps
as mainnet approaches.

While most of the gaps identified in this incident response were things the team
was already planning to put into place in the runup to mainnet, we've still approached
it as a full incident to practice our usual muscles, including creating
a full incident report with analysis and next steps. We've noted both the improvements
that we already had in the pipeline as well as any new observations in the
[Next Steps](#next-steps) section.

## Summary

`mezod`, the Mezo validator software, is a fork of an old version of [evmos],
from prior to a relicensing of the evmos software away from the [GNU General Public License][gpl]
to a more restrictive non-[copyleft] license. The fork predates certain security
fixes that evmos has applied, including some that allow for disagreements to
develop between the Cosmos and EVM layers of the system, creating opportunities
for bad accounting. When implementing `mezod`, the team was peripherally aware
of the classes of unpatched security bugs in the forked version of the evmos client,
and introduced commensurate safeguards to ensure that these bugs, even if exploited,
could not result in the opportunity to create more of the base asset (BTC) than
the chain had accounted for over a valid bridge.

Fixing this class of bug was one of the final elements of the pre-mainnet `mezod`
roadmap, and the team targeted an upgrade with related fixes to roll out to Matsnet
validators in the final weeks of March. Simultaneously, the team was doing testing
on [Tigris], the incentive system for Mezo mainnet. On March 12, 2025, around
15:57:18 UTC, a test transaction was included in a block that triggered a bug
within `mezod` that resulted in the creation of more BTC than the chain had
accounted for. This produced an immediate halt in block production as the validator
nodes all tripped the BTC supply safeguards.

Once the team dug in it quickly became apparent that the likely culprit was this
class of bug, and the release aimed for later was accelerated. Given the testnet
nature of the issue, we elected to let everyone get a good night's sleep rather
than the production approach of immediately focusing on fixing the system, extending
the downtime into the next day. Our initial hypothesis was that a cheeky user had
seen the recent open sourcing of the `mezod` repository, noticed the open bug, and
submitted an exploit transaction; exploration once the bug was fixed confirmed that
the issue was actually triggered by Tigris testing.

## Timeline

| Day | Time (UTC) | Event |
|-----|------------|-------|
| 2025-03-12 | 15:57 | Chain halts at block 3078794 |
| 2025-03-12 | 16:03 | First validator notices the issue and escalates to the development team |
| 2025-03-12 | 19:10 | The development team acknowledges the issue and starts the investigation |
| 2025-03-12 | 20:25 | The development team figures out the cause. A community announcement is published |
| 2025-03-13 | 11:25 | The v0.7.0-rc0 release fixing the issue is announced to the validators |
| 2025-03-13 | 13:58 | More than 66% of the validators apply the upgrade and the chain resumes |

## Origin of the incident

Versions v0.5.x (and lower) of `mezod` were vulnerable to a critical security
issue reported by Halborn, an auditing firm, in https://github.com/mezo-org/mezod/issues/401.
In short, state changes between EVM and Cosmos layers of the chain were not properly
propagated in all cases. The vulnerability allowed arbitrary minting/burning of testnet
BTC by calling the Mezo BTC ERC20 precompiled contract in specific ways.

**At this point, the Mezo chain client repository was still closed source. Moreover,
Matsnet testnet was the only existing network. Therefore, no real funds were ever at risk.**

Given the complex nature of the vulnerability, the development team approached the fix it in steps:

- In v0.6.0-rc0, the development team implemented a BTC supply safeguard whose goal was
  to halt the chain and prevent state corruption in case the mentioned vulnerability
  was exploited on Matsnet. The BTC supply safeguard was triggered on every block and
  asserted that the current BTC supply was equal to the difference between BTC minted
  and burned in the Mezo bridge. This effectively prevented arbitrary minting/burning
  of BTC beyond the Mezo bridge.
- In v0.7.0-rc0, the development team implemented a comprehensive fix that completely
  patched the mentioned security issue.

At the time of the incident, the Matsnet validators were still running on the v0.6.0
version line. The hard fork rolling out the v0.7.0 version line was tentatively
scheduled to happen in the week after the incident date.

Unfortunately, one of the Tigris test transactions executed on March 12 triggered
the vulnerability. As a result, the BTC supply safeguard detected the anomaly and
halted the chain.

## Response

In response to the incident, the development team conducted an investigation and confirmed
the BTC supply safeguard was violated due to the aforementioned security vulnerability being
triggered by a Tigris test transaction included in block 3078794.

The version line v0.7.0 containing the fix was almost ready to be released at this point,
so the development team finalized it and decided to roll it out immediately
among validators. Since the chain was halted, it was possible to do so without conducting
the usual on-chain upgrade/fork process that is executed under normal circumstances.

## Impact

As a result of the incident, the Matsnet testnet chain incurred 22 hours of downtime.
However, the on-chain state remained consistent. Neither test nor real funds were
ever at risk.

## Next steps

The development team will conduct the following actions to improve the incident
response practices. Most of them were already planned and put on the roadmap before
the incident. Those are:

- (Already planned) Build a set of emergency procedures, runbooks, escalation policies,
  and tooling that will cover chain halts and other failure scenarios that could happen
  on a live chain.
- (Already planned) Spin up a chain monitoring system with alerting capabilities
  to ensure failures are flagged early and escalated automatically.
- (Already planned) Conduct trial runs of the mentioned emergency procedures on
  Matsnet to practice muscles before the mainnet launch.  
- Build a status page acting as a frontend for the planned monitoring system
  that will show the health of the chain and uptime of specific validators.
- Optimize the release planning process to release changes more often,
  without unnecessary delays.  

[evmos]: https://evmos.org
[gpl]: https://www.gnu.org/licenses/gpl-3.0.html
[copyleft]: https://www.gnu.org/licenses/copyleft.en.html
[Tigris]: https://blog.mezo.org/mezo-the-2025-roadmap/#3-tigris
