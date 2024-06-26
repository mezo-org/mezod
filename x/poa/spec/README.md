<!--
order: 0
title: POA Overview
parent:
  title: "PoA"
-->

# `PoA`

## Abstract

The module enables a Cosmos-SDK based blockchain to use a Proof of Authority system to determine the validator set.

All validators in the system have equal voting power.

An initial validator set is defined in the genesis file.
Validators can be appended or kicked from the validator set through voting from the current validators.

A quorum percentage is defined in the parameters of the module. This quorum defines the number of approvals required to vote decision.

For example: if the quorum is 50% and the current validator set contains 10 validators. 5 validator approvals are required to accept a new candidate in the validator set.

## Contents

1. **[State](01_state.md)**
2. **[Messages](02_messages.md)**
3. **[End-Block ](03_end_block.md)**
4. **[Events](04_events.md)**
5. **[Parameters](05_params.md)**