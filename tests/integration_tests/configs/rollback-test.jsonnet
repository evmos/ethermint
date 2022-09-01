local config = import 'default.jsonnet';

config {
  'ethermint_9000-1'+: {
    validators: super.validators + [{
      name: 'fullnode',
    }],
  },
}
