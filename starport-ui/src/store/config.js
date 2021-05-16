import { env, starport, blocks, wallet,transfers } from '@starport/vuex'
import generated from './generated'
export default function init(store) {
	for (const moduleInit of Object.values(generated)) {
		moduleInit(store)
	}
	transfers(store)
	starport(store)
	blocks(store)
	env(store)
	wallet(store)
}
