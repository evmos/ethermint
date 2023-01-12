local config = import 'default.jsonnet';

config {
  'ethermint_9000-1'+: {
    'app-config'+: {
      'minimum-gas-prices': '100000000000aphoton',
    },
  },
}
