# Integration Tests

## `nix-shell`

`nix-shell` is used to achieve reproducible test environment setup. [Install Nix](https://nixos.org/download.html)

```
$ cd tests
$ nix-shell
<nix-shell> $
```

You can also use tools like [`direnv`](https://github.com/direnv/direnv) and [`lorri`](https://github.com/target/lorri/) together to get better dev UX.

After entered nix shell, you'll have `start-geth` and `start-ethermint` commands in `PATH`, both take data directory path as the only parameter, and start the devnet. You can further integration these commands into your test runner framework.

## Devnet Configuration

The devnet for geth and ethermint try to have identical configuration: 

- Both activated Berlin and older hardforks

- Have same genesis accounts and balances

  | Name      | Private key     | Address                                    | Balance     |
  | --------- | --------------- | ------------------------------------------ | ----------- |
  | validator | 826E47...A6EFF5 | 0x57f96e6b86cdefdb3d412547816a82e3e0ebf9d2 | 1photon     |
  | Community | 5D665F...043574 | 0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec | 10000photon |

### Ports

The geth devnet use default port `8485`, the ethermint devnet use `1317` for normal http rpc, `1318` for WebSocket.