local config = import 'default.jsonnet';

config {
  'ethermint_9000-1'+: {
    genesis+: {
      app_state+: {
        feemarket+: {
          params+: {
            no_base_fee: false,
            base_fee:: super.base_fee,
            initial_base_fee: super.base_fee,
          },
        },
      },
    },
  },
}
