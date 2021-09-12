<!--
order: 5
-->

# Validator FAQ

Check the FAQ for running a validator on Ethermint {synopsis}

## General Concepts

### What is a validator?

Ethermint is powered by [Tendermint](https://tendermint.com/docs/introduction/what-is-tendermint.html) Core, which relies on a set of validators to secure the network. Validators run a full node and participate in consensus by broadcasting votes which contain cryptographic signatures signed by their private key. Validators commit new blocks in the blockchain and receive revenue in exchange for their work. They also participate in on-procotol treasury governance by voting on governance proposals. A validator's voting influence is weighted according to their total stake.

### What is "staking"?

Ethermint is a public Proof-of-Stake (PoS) blockchain, meaning that validator's weight is determined by the amount of staking tokens (Photon) bonded as collateral. These staking tokens can be staked directly by the validator or delegated to them by Photon holders.

Any user in the system can declare its intention to become a validator by sending a [`create-validator`](#how-to-become-a-validator) transaction. From there, they become validators.

The weight (i.e. total stake or voting power) of a validator determines wether or not it is an active validator, and also how frequently this node will have to propose a block and how much revenue it will obtain. Initially, only the top 125 validators with the most weight will be active validators. If validators double-sign, or are frequently offline, they risk their staked tokens (including Photons delegated by users) being "slashed" by the protocol to penalize negligence and misbehavior.

### What is a full node?

A full node is a program that fully validates transactions and blocks of a blockchain. It is distinct from a light client node that only processes block headers and a small subset of transactions. Running a full node requires more resources than a light client but is necessary in order to be a validator. In practice, running a full-node only implies running a non-compromised and up-to-date version of the software with low network latency and without downtime.

Of course, it is possible and encouraged for any user to run full nodes even if they do not plan to be validators.

### What is a delegator?

Delegators are Photon holders who cannot, or do not want to run validator operations themselves. Users can delegate Photons to a validator and obtain a part of its revenue in exchange (for more detail on how revenue is distributed, see [What is the incentive to stake?](#what-is-the-incentive-to-stake) and [What is a validator's commission?](#what-is-a-validators-commission) sections below).

Because they share revenue with their validators, delegators also share responsibility. Should a validator misbehave, each of its delegators will be partially slashed in proportion to their stake. This is why delegators should perform due-diligence on validators before delegating, as well as diversifying by spreading their stake over multiple validators.

Delegators play a critical role in the system, as they are responsible for choosing validators. Be aware that being a delegator is not a passive role. Delegators are obligated to remain vigilant and actively monitor the actions of their validators, switching should they fail to act responsibly.

## Becoming a Validator

### How to become a validator?

Any participant in the network can signal their intent to become a validator by creating a validator and registering its validator profile. To do so, the candidate broadcasts a `create-validator` transaction, in which they must submit the following information:

- **Validator's PubKey**: Validator operators can have different accounts for validating and holding liquid funds. The PubKey submitted must be associated with the private key with which the validator intends to sign _prevotes_ and _precommits_.
- **Validator's Address**: `ethmvaloper1-` address. This is the address used to identify your validator publicly. The private key associated with this address is used to bond, unbond, and claim rewards.
- **Validator's name** (also known as the **moniker**)
- **Validator's website** _(optional)_
- **Validator's description** _(optional)_
- **Initial commission rate**: The commission rate on block provisions, block rewards and fees charged to delegators.
- **Maximum commission**: The maximum commission rate which this validator will be allowed to charge.
- **Commission change rate**: The maximum daily increase of the validator commission.
- **Minimum self-bond amount**: Minimum amount of Photon the validator needs to have bonded at all times. If the validator's self-bonded stake falls below this limit, its entire staking pool will be unbonded.
- **Initial self-bond amount**: Initial amount of Photon the validator wants to self-bond.

```bash
ethermintd tx staking create-validator
    --pubkey ethmvalconspub1zcjduepqs5s0vddx5m65h5ntjzwd0x8g3245rgrytpds4ds7vdtlwx06mcesmnkzly
    --amount "2aphoton"
    --from tmp
    --commission-rate="0.20"
    --commission-max-rate="1.00"
    --commission-max-change-rate="0.01"
    --min-self-delegation "1"
    --moniker "validator"
    --chain-id "ethermint_9000-1"
    --gas auto
    --node tcp://127.0.0.1:26647
```

Once a validator is created and registered, Photon holders can delegate Photons to it, effectively adding stake to its pool. The total stake of a validator is the sum of the Photon self-bonded by the validator's operator and the Photon bonded by external delegators.

**Only the top 125 validators with the most stake are considered the active validators**, becoming **bonded validators**. If ever a validator's total stake dips below the top 125, the validator loses its validator privileges (meaning that it won't generate rewards) and no longer serves as part of the active set (i.e doesn't participate in consensus), entering **unbonding mode** and eventually becomes **unbonded**.

## Validator keys and states

### What are the different types of keys?

In short, there are two types of keys:

- **Tendermint Key**: This is a unique key used to sign block hashes. It is associated with a public key `ethmvalconspub`.

  - Generated when the node is created with `ethermintd init`.
  - Get this value with `ethermintd tendermint show-validator`
    e.g. `ethmvalconspub1zcjduc3qcyj09qc03elte23zwshdx92jm6ce88fgc90rtqhjx8v0608qh5ssp0w94c`

- **Application keys**: These keys are created from the application and used to sign transactions. As a validator, you will probably use one key to sign staking-related transactions, and another key to sign oracle-related transactions. Application keys are associated with a public key `ethmpub-` and an address `ethm-`. Both are derived from account keys generated by `ethermintd keys add`.

::: warning
A validator's operator key is directly tied to an application key, but uses reserved prefixes solely for this purpose: `ethmvaloper` and `ethmvaloperpub`
:::

### What are the different states a validator can be in?

After a validator is created with a `create-validator` transaction, it can be in three states:

- `bonded`: Validator is in the active set and participates in consensus. Validator is earning rewards and can be slashed for misbehaviour.

- `unbonding`: Validator is not in the active set and does not participate in consensus. Validator is not earning rewards, but can still be slashed for misbehaviour. This is a transition state from `bonded` to `unbonded`. If validator does not send a `rebond` transaction while in `unbonding` mode, it will take three weeks for the state transition to complete.

- `unbonded`: Validator is not in the active set, and therefore not signing blocks. Unbonded validators cannot be slashed, but do not earn any rewards from their operation. It is still possible to delegate Photon to this validator. Un-delegating from an `unbonded` validator is immediate.

Delegators have the same state as their validator.

::: warning
Delegations are not necessarily bonded. Photon can be delegated and bonded, delegated and unbonding, delegated and unbonded, or liquid.
:::

### What is "self-bond"? How can I increase my "self-bond"?

The validator operator's "self-bond" refers to the amount of Photon stake delegated to itself. You can increase your self-bond by delegating more Photon to your validator account.

### Is there a faucet?

<!-- TODO: add link -->
If you want to obtain coins for the testnet, you can do so by using the faucet (link to be announced).

### Is there a minimum amount of Photon that must be staked to be an active (bonded) validator?

There is no minimum. The top 125 validators with the highest total stake (where `total stake = self-bonded stake + delegators stake`) are the active validators.

### How will delegators choose their validators?

Delegators are free to choose validators according to their own subjective criteria. That said, criteria anticipated to be important include:

- **Amount of self-bonded Photon:** Number of Photons a validator self-bonded to its staking pool. A validator with higher amount of self-bonded Photon has more skin in the game, making it more liable for its actions.

- **Amount of delegated Photons:** Total number of Photon delegated to a validator. A high stake shows that the community trusts this validator, but it also means that this validator is a bigger target for hackers. Validators are expected to become less and less attractive as their amount of delegated Photon grows. Bigger validators also increase the centralization of the network.

- **Commission rate:** Commission applied on revenue by validators before it is distributed to their delegators

- **Track record:** Delegators will likely look at the track record of the validators they plan to delegate to. This includes seniority, past votes on proposals, historical average uptime and how often the node was compromised.

Apart from these criteria, there will be a possibility for validators to signal a website address to complete their resume. Validators will need to build reputation one way or another to attract delegators. For example, it would be a good practice for validators to have their setup audited by third parties. Note though, that the Ethermint team will not approve or conduct any audit itself.

## Responsibilites

### Do validators need to be publicly identified?

No, they do not. Each delegator will value validators based on their own criteria. Validators will be able(and are advised) to register a website address when they nominate themselves so that they can advertise their operation as they see fit. Some delegators may prefer a website that clearly displays the team running the validator and their resume, while others might prefer anonymous validators with positive track records. Most likely both identified and anonymous validators will coexist in the validator set.

### What are the responsiblities of a validator?

Validators have three main responsibilities:

- **Be able to constantly run a correct version of the software:** validators need to make sure that their servers are always online and their private keys are not compromised.

- **Provide oversight and feedback on correct deployment of community pool funds:** the Ethermint protocol includes the a governance system for proposals to the facilitate adoption of its currencies. Validators are expected to hold budget executors to account to provide transparency and efficient use of funds.

Additionally, validators are expected to be active members of the community. They should always be up-to-date with the current state of the ecosystem so that they can easily adapt to any change.

### What does staking imply?

Staking Photon can be thought of as a safety deposit on validation activities. When a validator or a delegator wants to retrieve part or all of their deposit, they send an unbonding transaction. Then, Photon undergo a _three weeks unbonding period_ during which they are liable to being slashed for potential misbehaviors committed by the validator before the unbonding process started.

Validators, and by association delegators, receive block provisions, block rewards, and fee rewards. If a validator misbehaves, a certain portion of its total stake is slashed (the severity of the penalty depends on the type of misbehavior). This means that every user that bonded Photon to this validator gets penalized in proportion to its stake. Delegators are therefore incentivized to delegate to validators that they anticipate will function safely.

### Can a validator run away with its delegators' Photon?

By delegating to a validator, a user delegates staking power. The more staking power a validator has, the more weight it has in the consensus and processes. This does not mean that the validator has custody of its delegators' Photon. _By no means can a validator run away with its delegator's funds_.

Even though delegated funds cannot be stolen by their validators, delegators are still liable if their validators misbehave. In such case, each delegators' stake will be partially slashed in proportion to their relative stake.

### How often will a validator be chosen to propose the next block? Does it go up with the quantity of Photon staked?

The validator that is selected to mine the next block is called the **proposer**, the "leader" in the consensus for the round. Each proposer is selected deterministically, and the frequency of being chosen is equal to the relative total stake (where total stake = self-bonded stake + delegators stake) of the validator. For example, if the total bonded stake across all validators is 100 Photon, and a validator's total stake is 10 Photon, then this validator will be chosen 10% of the time as the proposer.

To understand more about the proposer selection process in Tendermint BFT consensus, read more [in their official docs](https://docs.tendermint.com/master/spec/reactors/consensus/proposer-selection.html).

## Incentives

### What is the incentive to stake?

Each member of a validator's staking pool earns different types of revenue:

- **Block rewards:** Native tokens of applications run by validators (e.g. Photons on Ethermint) are inflated to produce block provisions. These provisions exist to incentivize Photon holders to bond their stake, as non-bonded Photon will be diluted over time.
- **Transaction fees:** Ethermint maintains a whitelist of token that are accepted as fee payment. The initial fee token is the `photon`.

This total revenue is divided among validators' staking pools according to each validator's weight. Then, within each validator's staking pool the revenue is divided among delegators in proportion to each delegator's stake. A commission on delegators' revenue is applied by the validator before it is distributed.

### What is the incentive to run a validator ?

Validators earn proportionally more revenue than their delegators because of commissions.

Validators also play a major role in governance. If a delegator does not vote, they inherit the vote from their validator. This gives validators a major responsibility in the ecosystem.

### What is a validator's commission?

Revenue received by a validator's pool is split between the validator and its delegators. The validator can apply a commission on the part of the revenue that goes to its delegators. This commission is set as a percentage. Each validator is free to set its initial commission, maximum daily commission change rate and maximum commission. Ethermint enforces the parameter that each validator sets. These parameters can only be defined when initially declaring candidacy, and may only be constrained further after being declared.

### How are block provisions distributed?

Block provisions (rewards) are distributed proportionally to all validators relative to their total stake (voting power). This means that even though each validator gains Photons with each provision, all validators will still maintain equal weight.

Let us take an example where we have 10 validators with equal staking power and a commission rate of 1%. Let us also assume that the provision for a block is 1000 Photons and that each validator has 20% of self-bonded Photon. These tokens do not go directly to the proposer. Instead, they are evenly spread among validators. So now each validator's pool has 100 Photons. These 100 Photons will be distributed according to each participant's stake:

- Commission: `100*80%*1% = 0.8 Photons`
- Validator gets: `100\*20% + Commission = 20.8 Photons`
- All delegators get: `100\*80% - Commission = 79.2 Photons`

Then, each delegator can claim its part of the 79.2 Photons in proportion to their stake in the validator's staking pool. Note that the validator's commission is not applied on block provisions. Note that block rewards (paid in Photons) are distributed according to the same mechanism.

### How are fees distributed?

Fees are similarly distributed with the exception that the block proposer can get a bonus on the fees of the block it proposes if it includes more than the strict minimum of required precommits.

When a validator is selected to propose the next block, it must include at least ⅔ precommits for the previous block in the form of validator signatures. However, there is an incentive to include more than ⅔ precommits in the form of a bonus. The bonus is linear: it ranges from 1% if the proposer includes ⅔rd precommits (minimum for the block to be valid) to 5% if the proposer includes 100% precommits. Of course the proposer should not wait too long or other validators may timeout and move on to the next proposer. As such, validators have to find a balance between wait-time to get the most signatures and risk of losing out on proposing the next block. This mechanism aims to incentivize non-empty block proposals, better networking between validators as well as to mitigate censorship.

Let's take a concrete example to illustrate the aforementioned concept. In this example, there are 10 validators with equal stake. Each of them applies a 1% commission and has 20% of self-bonded Photon. Now comes a successful block that collects a total of 1005 Photons in fees. Let's assume that the proposer included 100% of the signatures in its block. It thus obtains the full bonus of 5%.

We have to solve this simple equation to find the reward $R$ for each validator:

$$9R ~ + ~ R ~ + ~ 5\%(R) ~ = ~ 1005 ~ \Leftrightarrow ~ R ~ = ~ 1005 ~/ ~10.05 ~ = ~ 100$$

- For the proposer validator:

  - The pool obtains $R ~ + ~ 5\%(R)$: 105 Photons
  - Commission: $105 ~ * ~ 80\% ~ * ~ 1\%$ = 0.84 Photons
  - Validator's reward: $105 ~ * ~ 20\% ~ + ~ Commission$ = 21.84 Photons
  - Delegators' rewards: $105 ~ * ~ 80\% ~ - ~ Commission$ = 83.16 Photons \(each delegator will be able to claim its portion of these rewards in proportion to their stake\)

  - The pool obtains $R$: 100 Photons
  - Commission: $100 ~ * ~ 80\% ~ * ~ 1\%$ = 0.8 Photons
  - Validator's reward: $100 ~ * ~ 20\% ~ + ~ Commission$ = 20.8 Photons
  - Delegators' rewards: $100 ~ * ~ 80\% ~ - ~ Commission$ = 79.2 Photons \(each delegator will be able to claim its portion of these rewards in proportion to their stake\)

### What are the slashing conditions?

If a validator misbehaves, its bonded stake along with its delegators' stake and will be slashed. The severity of the punishment depends on the type of fault. There are 3 main faults that can result in slashing of funds for a validator and its delegators:

- **Double-signing:** If someone reports on chain A that a validator signed two blocks at the same height on chain A and chain B, and if chain A and chain B share a common ancestor, then this validator will get slashed on chain A.

- **Downtime:** If a validator misses more than 95% of the last 10.000 blocks, they will get slashed by 0.01%.
- **Unavailability:** If a validator's signature has not been included in the last X blocks, the validator will get slashed by a marginal amount proportional to X. If X is above a certain limit Y, then the validator will get unbonded.
- **Non-voting:** If a validator did not vote on a proposal, its stake could receive a minor slash.

Note that even if a validator does not intentionally misbehave, it can still be slashed if its node crashes, looses connectivity, gets DDoSed, or if its private key is compromised.

### Do validators need to self-bond Photons

No, they do not. A validators total stake is equal to the sum of its own self-bonded stake and of its delegated stake. This means that a validator can compensate its low amount of self-bonded stake by attracting more delegators. This is why reputation is very important for validators.

Even though there is no obligation for validators to self-bond Photon, delegators should want their validator to have self-bonded Photon in their staking pool. In other words, validators should have skin-in-the-game.

In order for delegators to have some guarantee about how much skin-in-the-game their validator has, the latter can signal a minimum amount of self-bonded Photon. If a validator's self-bond goes below the limit that it predefined, this validator and all of its delegators will unbond.

### How to prevent concentration of stake in the hands of a few top validators?

For now the community is expected to behave in a smart and self-preserving way. When a mining pool in Bitcoin gets too much mining power the community usually stops contributing to that pool. Ethermint will rely on the same effect initially. In the future, other mechanisms will be deployed to smoothen this process as much as possible:

- **Penalty-free re-delegation:** This is to allow delegators to easily switch from one validator to another, in order to reduce validator stickiness.
- **UI warning:** Wallets can implement warnings that will be displayed to users if they want to delegate to a validator that already has a significant amount of staking power.

## Technical Requirements

### What are hardware requirements?

Validators should expect to provision one or more data center locations with redundant power, networking, firewalls, HSMs and servers.

We expect that a modest level of hardware specifications will be needed initially and that they might rise as network use increases. Participating in the testnet is the best way to learn more.

### What are software requirements?

In addition to running an Ethermint node, validators should develop monitoring, alerting and management solutions.

### What are bandwidth requirements?

Ethermint has the capacity for very high throughput compared to chains like Ethereum or Bitcoin.

As such, we recommend that the data center nodes only connect to trusted full nodes in the cloud or other validators that know each other socially. This relieves the data center node from the burden of mitigating denial-of-service attacks.

Ultimately, as the network becomes more used, one can realistically expect daily bandwidth on the order of several gigabytes.

### What does running a validator imply in terms of logistics?

A successful validator operation will require the efforts of multiple highly skilled individuals and continuous operational attention. This will be considerably more involved than running a bitcoin miner for instance.

### How to handle key management?

Validators should expect to run an HSM that supports ed25519 keys. Here are potential options:

- YubiHSM 2
- Ledger Nano S
- Ledger BOLOS SGX enclave
- Thales nShield support
- [Strangelove Horocrux](https://github.com/strangelove-ventures/horcrux)

The Ethermint team does not recommend one solution above the other. The community is encouraged to bolster the effort to improve HSMs and the security of key management.

### What can validators expect in terms of operations?

Running effective operation is the key to avoiding unexpectedly unbonding or being slashed. This includes being able to respond to attacks, outages, as well as to maintain security and isolation in your data center.

### What are the maintenance requirements?

Validators should expect to perform regular software updates to accommodate upgrades and bug fixes. There will inevitably be issues with the network early in its bootstrapping phase that will require substantial vigilance.

### How can validators protect themselves from Denial-of-Service attacks?

Denial-of-service attacks occur when an attacker sends a flood of internet traffic to an IP address to prevent the server at the IP address from connecting to the internet.

An attacker scans the network, tries to learn the IP address of various validator nodes and disconnect them from communication by flooding them with traffic.

One recommended way to mitigate these risks is for validators to carefully structure their network topology in a so-called sentry node architecture.

Validator nodes should only connect to full-nodes they trust because they operate them themselves or are run by other validators they know socially. A validator node will typically run in a data center. Most data centers provide direct links the networks of major cloud providers. The validator can use those links to connect to sentry nodes in the cloud. This shifts the burden of denial-of-service from the validator's node directly to its sentry nodes, and may require new sentry nodes be spun up or activated to mitigate attacks on existing ones.

Sentry nodes can be quickly spun up or change their IP addresses. Because the links to the sentry nodes are in private IP space, an internet based attacked cannot disturb them directly. This will ensure validator block proposals and votes always make it to the rest of the network.

It is expected that good operating procedures on that part of validators will completely mitigate these threats.

For more on sentry node architecture, see [this](https://forum.cosmos.network/t/sentry-node-architecture-overview/454).
