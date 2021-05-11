# Solidity tests

Increasingly difficult tests are provided:

- [Basic](./suites/basic): simple Counter example, for basic calls, transactions, and events
- [Initialize](./suites/initialize): initialization contract and tests from [aragonOS](https://github.com/aragon/aragonOS)
- [Initialize (Buidler)](./suites/initialize-buidler): initialization contract and tests from [aragonOS](https://github.com/aragon/aragonOS), using [buidler](https://buidler.dev/)
- [Proxy](./suites/proxy): depositable delegate proxy contract and tests from [aragonOS](https://github.com/aragon/aragonOS)
- [Staking](./suites/staking): Staking contracts and full test suite from [aragon/staking](http://github.com/aragon/staking)

### Quick start

**Prerequisite**: in the repo's root, run `make install` to install the `ethermintd` and `ethermintd` binaries. When done, come back to this directory.

**Prerequisite**: install the individual solidity packages. They're set up as individual reops in a yarn monorepo workspace. Install them all via `yarn install`.

To run the tests, start three terminals (or two, if you run `ethermintd` with `&`).

In the first, run `ethermintd`:

```sh
./init-test-node.sh
```

In the second, run `ethermintd` as mentioned in the script's output:

```sh
ethermintd rest-server --laddr "tcp://localhost:8545" --unlock-key localkey,user1,user2 --chain-id "ethermint-1337" --trace --wsport 8546
```

You will now have three ethereum accounts unlocked in the test node:

- `0x3b7252d007059ffc82d16d022da3cbf9992d2f70` (Validator)
- `0xddd64b4712f7c8f1ace3c145c950339eddaf221d` (User 1)
- `0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0` (user 2)

From here, in your other available terminal, go into any of the tests and run `yarn test-ethermint`. You should see `ethermintd` accepting transactions and producing blocks. You should be able to query for any transaction via:

- `ethermintd query tx <cosmos-sdk tx>`
- `curl localhost:8545 -H "Content-Type:application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["<ethereum tx>"],"id":1}'`

And obviously more, via the Ethereum JSON-RPC API).

When in doubt, you can also run the tests against a Ganache instance via `yarn test-ganache`, to make sure they are behaving correctly.

### Test node

The [`init-test-node.sh`](./init-test-node.sh) script sets up ethermint with the following accounts:

- `eth18de995q8qk0leqk3d5pzmg7tlxvj6tmsku084d` (Validator)
  - `0x3b7252d007059ffc82d16d022da3cbf9992d2f70`
- `eth1mhtyk3cj7ly0rt8rc9zuj5pnnmw67gsapygwyq` (User 1)
  - `0xddd64b4712f7c8f1ace3c145c950339eddaf221d`
- `eth1pa20g7lehr330vs5ent20slr3wyne4lsy8qae3` (user 2)
  - `0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0`

Each with roughly 100 ETH available (1e18 photon).

Running `ethermintd list keys` should output:

```json
[
  {
    "name": "localkey",
    "type": "local",
    "address": "eth18de995q8qk0leqk3d5pzmg7tlxvj6tmsku084d",
    "pubkey": "ethpub1pfqnmk6pq3ycjs34vv4n6rkty89f6m02qcsal3ecdzn7a3uunx0e5ly0846pzg903hxf2zp5gq4grh8jcatcemfrscdfl797zhg5crkcsx43gujzppge3n"
  },
  {
    "name": "user1",
    "type": "local",
    "address": "eth1mhtyk3cj7ly0rt8rc9zuj5pnnmw67gsapygwyq",
    "pubkey": "ethpub1pfqnmk6pq3wrkx6lh7uug8ss0thggact3n49m5gkmpca4vylldpur5qrept57e0rrxfmeq5mp5xt3cyf4kys53qcv66qxttv970das69hlpkf8cnyd2a2x"
  },
  {
    "name": "user2",
    "type": "local",
    "address": "eth1pa20g7lehr330vs5ent20slr3wyne4lsy8qae3",
    "pubkey": "ethpub1pfqnmk6pq3art9y45zw5ntyktt2qrt0skmsl0ux9qwk8458ed3d8sgnrs99zlgvj3rt2vggvkh0x56hffugwsyddwqla48npx46pglgs6xhcqpall58tgn"
  }
]
```

And running:

```sh
curl localhost:8545 -H "Content-Type:application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}'
```

Should output:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    "0x3b7252d007059ffc82d16d022da3cbf9992d2f70",
    "0xddd64b4712f7c8f1ace3c145c950339eddaf221d",
    "0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0"
  ]
}
```
