import { txClient, queryClient } from './module'
// @ts-ignore
import { SpVuexError } from '@starport/vuex'

import { GenesisAccount } from "./module/types/ethermint/evm/v1alpha1/genesis"
import { Params } from "./module/types/ethermint/evm/v1alpha1/evm"
import { ChainConfig } from "./module/types/ethermint/evm/v1alpha1/evm"
import { State } from "./module/types/ethermint/evm/v1alpha1/evm"
import { TransactionLogs } from "./module/types/ethermint/evm/v1alpha1/evm"
import { Log } from "./module/types/ethermint/evm/v1alpha1/evm"
import { TxReceipt } from "./module/types/ethermint/evm/v1alpha1/evm"
import { TxResult } from "./module/types/ethermint/evm/v1alpha1/evm"
import { TxData } from "./module/types/ethermint/evm/v1alpha1/evm"
import { BytesList } from "./module/types/ethermint/evm/v1alpha1/evm"
import { AccessTuple } from "./module/types/ethermint/evm/v1alpha1/evm"
import { ExtensionOptionsEthereumTx } from "./module/types/ethermint/evm/v1alpha1/tx"
import { ExtensionOptionsWeb3Tx } from "./module/types/ethermint/evm/v1alpha1/tx"


async function initTxClient(vuexGetters) {
	return await txClient(vuexGetters['common/wallet/signer'], {
		addr: vuexGetters['common/env/apiTendermint']
	})
}

async function initQueryClient(vuexGetters) {
	return await queryClient({
		addr: vuexGetters['common/env/apiCosmos']
	})
}

function getStructure(template) {
	let structure = { fields: [] }
	for (const [key, value] of Object.entries(template)) {
		let field: any = {}
		field.name = key
		field.type = typeof value
		structure.fields.push(field)
	}
	return structure
}

const getDefaultState = () => {
	return {
        Account: {},
        CosmosAccount: {},
        Balance: {},
        Storage: {},
        Code: {},
        TxLogs: {},
        TxReceipt: {},
        TxReceiptsByBlockHeight: {},
        TxReceiptsByBlockHash: {},
        BlockLogs: {},
        BlockBloom: {},
        Params: {},
        StaticCall: {},
        
        _Structure: {
            GenesisAccount: getStructure(GenesisAccount.fromPartial({})),
            Params: getStructure(Params.fromPartial({})),
            ChainConfig: getStructure(ChainConfig.fromPartial({})),
            State: getStructure(State.fromPartial({})),
            TransactionLogs: getStructure(TransactionLogs.fromPartial({})),
            Log: getStructure(Log.fromPartial({})),
            TxReceipt: getStructure(TxReceipt.fromPartial({})),
            TxResult: getStructure(TxResult.fromPartial({})),
            TxData: getStructure(TxData.fromPartial({})),
            BytesList: getStructure(BytesList.fromPartial({})),
            AccessTuple: getStructure(AccessTuple.fromPartial({})),
            ExtensionOptionsEthereumTx: getStructure(ExtensionOptionsEthereumTx.fromPartial({})),
            ExtensionOptionsWeb3Tx: getStructure(ExtensionOptionsWeb3Tx.fromPartial({})),
            
		},
		_Subscriptions: new Set(),
	}
}

// initial state
const state = getDefaultState()

export default {
	namespaced: true,
	state,
	mutations: {
		RESET_STATE(state) {
			Object.assign(state, getDefaultState())
		},
		QUERY(state, { query, key, value }) {
			state[query][JSON.stringify(key)] = value
		},
		SUBSCRIBE(state, subscription) {
			state._Subscriptions.add(subscription)
		},
		UNSUBSCRIBE(state, subscription) {
			state._Subscriptions.delete(subscription)
		}
	},
	getters: {
        getAccount: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Account[JSON.stringify(params)] ?? {}
		},
        getCosmosAccount: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.CosmosAccount[JSON.stringify(params)] ?? {}
		},
        getBalance: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Balance[JSON.stringify(params)] ?? {}
		},
        getStorage: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Storage[JSON.stringify(params)] ?? {}
		},
        getCode: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Code[JSON.stringify(params)] ?? {}
		},
        getTxLogs: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.TxLogs[JSON.stringify(params)] ?? {}
		},
        getTxReceipt: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.TxReceipt[JSON.stringify(params)] ?? {}
		},
        getTxReceiptsByBlockHeight: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.TxReceiptsByBlockHeight[JSON.stringify(params)] ?? {}
		},
        getTxReceiptsByBlockHash: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.TxReceiptsByBlockHash[JSON.stringify(params)] ?? {}
		},
        getBlockLogs: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.BlockLogs[JSON.stringify(params)] ?? {}
		},
        getBlockBloom: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.BlockBloom[JSON.stringify(params)] ?? {}
		},
        getParams: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Params[JSON.stringify(params)] ?? {}
		},
        getStaticCall: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.StaticCall[JSON.stringify(params)] ?? {}
		},
        
		getTypeStructure: (state) => (type) => {
			return state._Structure[type].fields
		}
	},
	actions: {
		init({ dispatch, rootGetters }) {
			console.log('init')
			if (rootGetters['common/env/client']) {
				rootGetters['common/env/client'].on('newblock', () => {
					dispatch('StoreUpdate')
				})
			}
		},
		resetState({ commit }) {
			commit('RESET_STATE')
		},
		unsubscribe({ commit }, subscription) {
			commit('UNSUBSCRIBE', subscription)
		},
		async StoreUpdate({ state, dispatch }) {
			state._Subscriptions.forEach((subscription) => {
				dispatch(subscription.action, subscription.payload)
			})
		},
		async QueryAccount({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryAccount( key.address,  query)).data:(await (await initQueryClient(rootGetters)).queryAccount( key.address )).data
				
				commit('QUERY', { query: 'Account', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryAccount', payload: { options: { all }, params: {...key},query }})
				return getters['getAccount']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryAccount', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryCosmosAccount({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryCosmosAccount( key.address,  query)).data:(await (await initQueryClient(rootGetters)).queryCosmosAccount( key.address )).data
				
				commit('QUERY', { query: 'CosmosAccount', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryCosmosAccount', payload: { options: { all }, params: {...key},query }})
				return getters['getCosmosAccount']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryCosmosAccount', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryBalance({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryBalance( key.address,  query)).data:(await (await initQueryClient(rootGetters)).queryBalance( key.address )).data
				
				commit('QUERY', { query: 'Balance', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryBalance', payload: { options: { all }, params: {...key},query }})
				return getters['getBalance']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryBalance', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryStorage({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryStorage( key.address,  key.key,  query)).data:(await (await initQueryClient(rootGetters)).queryStorage( key.address ,  key.key )).data
				
				commit('QUERY', { query: 'Storage', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryStorage', payload: { options: { all }, params: {...key},query }})
				return getters['getStorage']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryStorage', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryCode({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryCode( key.address,  query)).data:(await (await initQueryClient(rootGetters)).queryCode( key.address )).data
				
				commit('QUERY', { query: 'Code', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryCode', payload: { options: { all }, params: {...key},query }})
				return getters['getCode']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryCode', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryTxLogs({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryTxLogs( key.hash,  query)).data:(await (await initQueryClient(rootGetters)).queryTxLogs( key.hash )).data
				
				commit('QUERY', { query: 'TxLogs', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryTxLogs', payload: { options: { all }, params: {...key},query }})
				return getters['getTxLogs']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryTxLogs', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryTxReceipt({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryTxReceipt( key.hash,  query)).data:(await (await initQueryClient(rootGetters)).queryTxReceipt( key.hash )).data
				
				commit('QUERY', { query: 'TxReceipt', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryTxReceipt', payload: { options: { all }, params: {...key},query }})
				return getters['getTxReceipt']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryTxReceipt', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryTxReceiptsByBlockHeight({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryTxReceiptsByBlockHeight( key.height,  query)).data:(await (await initQueryClient(rootGetters)).queryTxReceiptsByBlockHeight( key.height )).data
				
				commit('QUERY', { query: 'TxReceiptsByBlockHeight', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryTxReceiptsByBlockHeight', payload: { options: { all }, params: {...key},query }})
				return getters['getTxReceiptsByBlockHeight']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryTxReceiptsByBlockHeight', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryTxReceiptsByBlockHash({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryTxReceiptsByBlockHash( key.hash,  query)).data:(await (await initQueryClient(rootGetters)).queryTxReceiptsByBlockHash( key.hash )).data
				
				commit('QUERY', { query: 'TxReceiptsByBlockHash', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryTxReceiptsByBlockHash', payload: { options: { all }, params: {...key},query }})
				return getters['getTxReceiptsByBlockHash']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryTxReceiptsByBlockHash', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryBlockLogs({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryBlockLogs( key.hash,  query)).data:(await (await initQueryClient(rootGetters)).queryBlockLogs( key.hash )).data
				
				commit('QUERY', { query: 'BlockLogs', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryBlockLogs', payload: { options: { all }, params: {...key},query }})
				return getters['getBlockLogs']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryBlockLogs', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryBlockBloom({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryBlockBloom( query)).data:(await (await initQueryClient(rootGetters)).queryBlockBloom()).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryBlockBloom({...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'BlockBloom', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryBlockBloom', payload: { options: { all }, params: {...key},query }})
				return getters['getBlockBloom']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryBlockBloom', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryParams({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryParams( query)).data:(await (await initQueryClient(rootGetters)).queryParams()).data
				
				commit('QUERY', { query: 'Params', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryParams', payload: { options: { all }, params: {...key},query }})
				return getters['getParams']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryParams', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryStaticCall({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryStaticCall( query)).data:(await (await initQueryClient(rootGetters)).queryStaticCall()).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryStaticCall({...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'StaticCall', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryStaticCall', payload: { options: { all }, params: {...key},query }})
				return getters['getStaticCall']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryStaticCall', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		
		async sendMsgEthereumTx({ rootGetters }, { value, fee, memo }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgEthereumTx(value)
				const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], {fee: { amount: fee, 
  gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgEthereumTx:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgEthereumTx:Send', 'Could not broadcast Tx.')
				}
			}
		},
		
		async MsgEthereumTx({ rootGetters }, { value }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgEthereumTx(value)
				return msg
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgEthereumTx:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgEthereumTx:Create', 'Could not create message.')
				}
			}
		},
		
	}
}
