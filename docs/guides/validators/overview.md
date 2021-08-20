<!--
order: 1
-->

# Overview

## Introduction

Ethermint is based on [Tendermint](https://github.com/tendermint/tendermint/tree/master/docs/introduction), which relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes which contain cryptographic signatures signed by each validator's private key.

Validator candidates can bond their own staking tokens (Photons) and have the tokens "delegated", or staked, to them by Photon holders. Ethermint will have 125 validators, but over time this will increase to 300 validators according to a predefined schedule. The validators are determined by who has the most stake delegated to them — the top 125 validator candidates with the most stake will become Cosmos validators.

Validators and their delegators will earn Atoms as block provisions and tokens as transaction fees through execution of the Tendermint consensus protocol. Initially, transaction fees will be paid in Atoms but in the future, any token in the Cosmos ecosystem will be valid as fee tender if it is whitelisted by governance. Note that validators can set commission on the fees their delegators receive as additional incentive.

If validators double sign, are frequently offline or do not participate in governance, their staked Atoms (including Atoms of users that delegated to them) can be slashed. The penalty depends on the severity of the violation.

## Hardware

Validators should set up a physical operation secured with restricted access. A good starting place, for example, would be co-locating in secure data centers.

Validators should expect to equip their datacenter location with redundant power, connectivity, and storage backups. Expect to have several redundant networking boxes for fiber, firewall and switching and then small servers with redundant hard drive and failover. Hardware can be on the low end of datacenter gear to start out with.

We anticipate that network requirements will be low initially. Then bandwidth, CPU and memory requirements will rise as the network grows. Large hard drives are recommended for storing years of blockchain history.

## Set Up a Website

Set up a dedicated validator's website and signal your intention to become a validator on Discord. This is important since delegators will want to have information about the entity they are delegating their Photons to.

## Seek Legal Advice

Seek legal advice if you intend to run a Validator.

## Community

Discuss the finer details of being a validator on our community chat and forum:

* [Validator Forum](https://forum.cosmos.network/c/validating)
