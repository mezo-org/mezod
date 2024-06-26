<!--
order: 4
-->

# Events

The poa module emits the following events:

## Handlers

### MsgSubmitApplication

**If Quorum > 0%:**

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| submit_application | candidate     | {validatorAddress} |
| submit_application | module     | poa |

**If Quorum = 0%:**

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| append_validator | candidate     | {validatorAddress} |
| append_validator | module     | poa |


### MsgProposeKick

**If Quorum > 0%:**

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| propose_kick | validator     | {validatorAddress} |
| propose_kick | proposer     | {validatorAddress} |
| propose_kick | module     | poa |


**If Quorum = 0%:**

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| kick_validator | validator     | {validatorAddress} |
| kick_validator | module     | poa |


### MsgLeaveValidatorSet

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| leave_validator_set | validator     | {validatorAddress} |
| leave_validator_set | module     | poa |


### MsgVote

#### Approve application

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| approve_application | voter     | {validatorAddress} |
| approve_application | candidate     | {validatorAddress} |
| approve_application | module     | poa |


**If Quorum reached:**

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| append_validator | candidate     | {validatorAddress} |
| append_validator | module     | poa |


#### Reject application

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| reject_application | voter     | {validatorAddress} |
| reject_application | candidate     | {validatorAddress} |
| reject_application | module     | poa |


**If Quorum reached for rejection:**

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| reject_validator | candidate     | {validatorAddress} |
| reject_validator | module     | poa |


#### Approve kick proposal

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| approve_kick_proposal | voter     | {validatorAddress} |
| approve_kick_proposal | validator     | {validatorAddress} |
| approve_kick_proposal | module     | poa |


**If Quorum reached:**

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| kick_validator | validator     | {validatorAddress} |
| kick_validator | module     | poa |


#### Reject kick proposal

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| reject_kick_proposal | voter     | {validatorAddress} |
| reject_kick_proposal | validator     | {validatorAddress} |
| reject_kick_proposal | module     | poa |


**If Quorum reached:**

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| keep_validator | validator     | {validatorAddress} |
| keep_validator | module     | poa |

