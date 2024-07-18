<!--
order: 0
title: POA Overview
parent:
  title: "PoA"
-->

# `PoA`

## Abstract

The module enables a Cosmos-SDK based blockchain to use a Proof of Authority 
system to determine the validator set.

An initial validator set is defined in the genesis file. Subsequent
validators submit candidate applications to join the validator set. The
module's owner has an authority to approve these applications.

A validator can voluntarily leave the validator set or be kicked out by the
module's owner.

All validators in the system have equal voting power.

The module's owner can transfer the ownership to another account in a 2-step
process. The new owner must accept the ownership transfer before the transfer
is completed. The initial owner must be defined in the genesis file.

## Contents

1. **[State](01_state.md)**
2. **[Start-Block](02_start_block.md)**
3. **[End-Block](03_end_block.md)**
4. **[Parameters](04_params.md)**
