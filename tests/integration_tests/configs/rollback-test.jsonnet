local config = import 'default.jsonnet';

config {
  'cronos_777-1'+: {
    validators: super.validators + [{
      name: 'fullnode',
    }],
    genesis+: {
      app_state+: {
        feemarket: {
          params: {
            no_base_fee: true,
          },
        },
      },
    },
  },
}
