<!--
order: 4
-->

# Run a Validator

Configure a validator node to propose blocks and earn staking rewards {synopsis}

## Pre-requisite Readings

- [Installation](./installation.md) {prereq}
- [Run a Full Node](./run_node.md) {prereq}

## What is a Validator?

[Validators](https://hub.cosmos.network/master/validators/overview.html) are responsible for committing new blocks to the blockchain through voting. A validator's stake is slashed if they become unavailable or sign blocks at the same height. Please read about [Sentry Node Architecture](https://hub.cosmos.network/master/validators/validator-faq.html#how-can-validators-protect-themselves-from-denial-of-service-attacks) to protect your node from DDOS attacks and to ensure high-availability.

::: danger Warning
If you want to become a validator for `mainnet`, you should [research security](https://hub.cosmos.network/master/validators/security.html).
:::

You may want to skip the next section if you have already set up a [full node](../emint-tutorials/join-mainnet.md).

## Create Your Validator

Your `cosmosvalconspub` consensus public key fron tendermint can be used to create a new validator by staking tokens. You can find your validator pubkey by running:

```bash
ethermintd tendermint show-validator
```

To create your validator, just use the following command:

```bash
ethermintd tx staking create-validator \
  --amount=1000000aphoton \
  --pubkey=$(ethermintd tendermint show-validator) \
  --moniker=<ethermint_validator> \
  --chain-id=<chain_id> \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
  --gas-prices="0.025uatom" \
  --from=<key_name>
```

::: tip
When specifying commission parameters, the `commission-max-change-rate` is used to measure % _point_ change over the `commission-rate`. E.g. 1% to 2% is a 100% rate increase, but only 1 percentage point.
:::

::: tip
`Min-self-delegation` is a stritly positive integer that represents the minimum amount of self-delegated voting power your validator must always have. A `min-self-delegation` of 1 means your validator will never have a self-delegation lower than `1000000aphoton`
:::

You can confirm that you are in the validator set by using a third party explorer.

## Genesis Transactions

A genesis transaction (aka `gentx`) is a JSON file carrying a self-delegation from a validator. All genesis transactions are collected by a genesis coordinator and validated against an initial `genesis.json` file.

A `gentx` does three things:

1. Makes the `validator` account you created into a validator operator account (i.e. the account that controls the validator).
2. Self-delegates the provided `amount` of staking tokens.
3. Link the operator account with a Tendermint node pubkey that will be used for signing blocks. If no `--pubkey` flag is provided, it defaults to the local node pubkey created via the `ethermintd init` command above.

If you want to participate in genesis as a validator, you need to justify that
you have some stake at genesis, create one (or multiple) transactions to bond this stake to your validator address, and include this transaction in the genesis file.

Your `cosmosvalconspub`, as shown on the section above, can be used to create a validator transaction on genesis as well.

Next, craft your `ethermintd gentx` command:

::: tip
When specifying commission parameters, the `commission-max-change-rate` is used to measure % _point_ change over the `commission-rate`. E.g. 1% to 2% is a 100% rate increase, but only 1 percentage point.
:::

```bash
ethermintd gentx \
  --amount <amount_of_delegation_uatom> \
  --commission-rate <commission_rate> \
  --commission-max-rate <commission_max_rate> \
  --commission-max-change-rate <commission_max_change_rate> \
  --pubkey $(ethermintd tendermint show-validator) \
  --name $KEY
```

::: tip
For more on `gentx`, use the help flag: `ethermintd gentx -h`
:::

## Confirm Your Validator is Running

Your validator is active if the following command returns anything:

```bash
ethermintd query tendermint-validator-set | grep "$(ethermintd tendermint show-validator)"
```

You should now see your validator in one of the block explorers. You are looking for the `bech32`
encoded `address` in the `~/.ethermintd/config/priv_validator.json` file.

::: tip
To be in the validator set, you need to have more total voting power than the 100th validator.
:::

## Halt Your Validator Node

When attempting to perform routine maintenance or planning for an upcoming coordinated
upgrade, it can be useful to have your validator systematically and gracefully halt the chain and shutdown the node.

You can achieve this by setting one of the following flags during when using the `ethermintd start` command:

- `--halt-height`: to the block height at which to shutdown the node
- `--halt-time`: to the minimum block time (in Unix seconds) at which to shutdown the node

The node will stop processing blocks with a zero exit code at that given height/time after
committing the block.

## Next {hide}

Start and connect a [client](./clients.md) to a running network {hide}
