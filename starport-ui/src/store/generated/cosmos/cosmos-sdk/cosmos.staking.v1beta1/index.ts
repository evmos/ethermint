import { txClient, queryClient } from './module'
// @ts-ignore
import { SpVuexError } from '@starport/vuex'

import { HistoricalInfo } from "./module/types/cosmos/staking/v1beta1/staking"
import { CommissionRates } from "./module/types/cosmos/staking/v1beta1/staking"
import { Commission } from "./module/types/cosmos/staking/v1beta1/staking"
import { Description } from "./module/types/cosmos/staking/v1beta1/staking"
import { Validator } from "./module/types/cosmos/staking/v1beta1/staking"
import { ValAddresses } from "./module/types/cosmos/staking/v1beta1/staking"
import { DVPair } from "./module/types/cosmos/staking/v1beta1/staking"
import { DVPairs } from "./module/types/cosmos/staking/v1beta1/staking"
import { DVVTriplet } from "./module/types/cosmos/staking/v1beta1/staking"
import { DVVTriplets } from "./module/types/cosmos/staking/v1beta1/staking"
import { Delegation } from "./module/types/cosmos/staking/v1beta1/staking"
import { UnbondingDelegation } from "./module/types/cosmos/staking/v1beta1/staking"
import { UnbondingDelegationEntry } from "./module/types/cosmos/staking/v1beta1/staking"
import { RedelegationEntry } from "./module/types/cosmos/staking/v1beta1/staking"
import { Redelegation } from "./module/types/cosmos/staking/v1beta1/staking"
import { Params } from "./module/types/cosmos/staking/v1beta1/staking"
import { DelegationResponse } from "./module/types/cosmos/staking/v1beta1/staking"
import { RedelegationEntryResponse } from "./module/types/cosmos/staking/v1beta1/staking"
import { RedelegationResponse } from "./module/types/cosmos/staking/v1beta1/staking"
import { Pool } from "./module/types/cosmos/staking/v1beta1/staking"
import { LastValidatorPower } from "./module/types/cosmos/staking/v1beta1/genesis"


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
        Validators: {},
        Validator: {},
        ValidatorDelegations: {},
        ValidatorUnbondingDelegations: {},
        Delegation: {},
        UnbondingDelegation: {},
        DelegatorDelegations: {},
        DelegatorUnbondingDelegations: {},
        Redelegations: {},
        DelegatorValidators: {},
        DelegatorValidator: {},
        HistoricalInfo: {},
        Pool: {},
        Params: {},
        
        _Structure: {
            HistoricalInfo: getStructure(HistoricalInfo.fromPartial({})),
            CommissionRates: getStructure(CommissionRates.fromPartial({})),
            Commission: getStructure(Commission.fromPartial({})),
            Description: getStructure(Description.fromPartial({})),
            Validator: getStructure(Validator.fromPartial({})),
            ValAddresses: getStructure(ValAddresses.fromPartial({})),
            DVPair: getStructure(DVPair.fromPartial({})),
            DVPairs: getStructure(DVPairs.fromPartial({})),
            DVVTriplet: getStructure(DVVTriplet.fromPartial({})),
            DVVTriplets: getStructure(DVVTriplets.fromPartial({})),
            Delegation: getStructure(Delegation.fromPartial({})),
            UnbondingDelegation: getStructure(UnbondingDelegation.fromPartial({})),
            UnbondingDelegationEntry: getStructure(UnbondingDelegationEntry.fromPartial({})),
            RedelegationEntry: getStructure(RedelegationEntry.fromPartial({})),
            Redelegation: getStructure(Redelegation.fromPartial({})),
            Params: getStructure(Params.fromPartial({})),
            DelegationResponse: getStructure(DelegationResponse.fromPartial({})),
            RedelegationEntryResponse: getStructure(RedelegationEntryResponse.fromPartial({})),
            RedelegationResponse: getStructure(RedelegationResponse.fromPartial({})),
            Pool: getStructure(Pool.fromPartial({})),
            LastValidatorPower: getStructure(LastValidatorPower.fromPartial({})),
            
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
        getValidators: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Validators[JSON.stringify(params)] ?? {}
		},
        getValidator: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Validator[JSON.stringify(params)] ?? {}
		},
        getValidatorDelegations: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.ValidatorDelegations[JSON.stringify(params)] ?? {}
		},
        getValidatorUnbondingDelegations: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.ValidatorUnbondingDelegations[JSON.stringify(params)] ?? {}
		},
        getDelegation: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Delegation[JSON.stringify(params)] ?? {}
		},
        getUnbondingDelegation: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.UnbondingDelegation[JSON.stringify(params)] ?? {}
		},
        getDelegatorDelegations: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.DelegatorDelegations[JSON.stringify(params)] ?? {}
		},
        getDelegatorUnbondingDelegations: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.DelegatorUnbondingDelegations[JSON.stringify(params)] ?? {}
		},
        getRedelegations: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Redelegations[JSON.stringify(params)] ?? {}
		},
        getDelegatorValidators: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.DelegatorValidators[JSON.stringify(params)] ?? {}
		},
        getDelegatorValidator: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.DelegatorValidator[JSON.stringify(params)] ?? {}
		},
        getHistoricalInfo: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.HistoricalInfo[JSON.stringify(params)] ?? {}
		},
        getPool: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Pool[JSON.stringify(params)] ?? {}
		},
        getParams: (state) => (params = {}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Params[JSON.stringify(params)] ?? {}
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
		async QueryValidators({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryValidators( query)).data:(await (await initQueryClient(rootGetters)).queryValidators()).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryValidators({...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'Validators', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryValidators', payload: { options: { all }, params: {...key},query }})
				return getters['getValidators']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryValidators', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryValidator({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryValidator( key.validator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryValidator( key.validator_addr )).data
				
				commit('QUERY', { query: 'Validator', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryValidator', payload: { options: { all }, params: {...key},query }})
				return getters['getValidator']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryValidator', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryValidatorDelegations({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryValidatorDelegations( key.validator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryValidatorDelegations( key.validator_addr )).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryValidatorDelegations( key.validator_addr, {...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'ValidatorDelegations', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryValidatorDelegations', payload: { options: { all }, params: {...key},query }})
				return getters['getValidatorDelegations']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryValidatorDelegations', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryValidatorUnbondingDelegations({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryValidatorUnbondingDelegations( key.validator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryValidatorUnbondingDelegations( key.validator_addr )).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryValidatorUnbondingDelegations( key.validator_addr, {...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'ValidatorUnbondingDelegations', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryValidatorUnbondingDelegations', payload: { options: { all }, params: {...key},query }})
				return getters['getValidatorUnbondingDelegations']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryValidatorUnbondingDelegations', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryDelegation({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryDelegation( key.validator_addr,  key.delegator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryDelegation( key.validator_addr ,  key.delegator_addr )).data
				
				commit('QUERY', { query: 'Delegation', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDelegation', payload: { options: { all }, params: {...key},query }})
				return getters['getDelegation']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryDelegation', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryUnbondingDelegation({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryUnbondingDelegation( key.validator_addr,  key.delegator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryUnbondingDelegation( key.validator_addr ,  key.delegator_addr )).data
				
				commit('QUERY', { query: 'UnbondingDelegation', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryUnbondingDelegation', payload: { options: { all }, params: {...key},query }})
				return getters['getUnbondingDelegation']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryUnbondingDelegation', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryDelegatorDelegations({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryDelegatorDelegations( key.delegator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryDelegatorDelegations( key.delegator_addr )).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryDelegatorDelegations( key.delegator_addr, {...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'DelegatorDelegations', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDelegatorDelegations', payload: { options: { all }, params: {...key},query }})
				return getters['getDelegatorDelegations']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryDelegatorDelegations', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryDelegatorUnbondingDelegations({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryDelegatorUnbondingDelegations( key.delegator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryDelegatorUnbondingDelegations( key.delegator_addr )).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryDelegatorUnbondingDelegations( key.delegator_addr, {...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'DelegatorUnbondingDelegations', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDelegatorUnbondingDelegations', payload: { options: { all }, params: {...key},query }})
				return getters['getDelegatorUnbondingDelegations']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryDelegatorUnbondingDelegations', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryRedelegations({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryRedelegations( key.delegator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryRedelegations( key.delegator_addr )).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryRedelegations( key.delegator_addr, {...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'Redelegations', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryRedelegations', payload: { options: { all }, params: {...key},query }})
				return getters['getRedelegations']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryRedelegations', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryDelegatorValidators({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryDelegatorValidators( key.delegator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryDelegatorValidators( key.delegator_addr )).data
				
				while (all && (<any> value).pagination && (<any> value).pagination.nextKey!=null) {
					let next_values=(await (await initQueryClient(rootGetters)).queryDelegatorValidators( key.delegator_addr, {...query, 'pagination.key':(<any> value).pagination.nextKey})).data
					for (let prop of Object.keys(next_values)) {
						if (Array.isArray(next_values[prop])) {
							value[prop]=[...value[prop], ...next_values[prop]]
						}else{
							value[prop]=next_values[prop]
						}
					}
				}
				
				commit('QUERY', { query: 'DelegatorValidators', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDelegatorValidators', payload: { options: { all }, params: {...key},query }})
				return getters['getDelegatorValidators']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryDelegatorValidators', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryDelegatorValidator({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryDelegatorValidator( key.delegator_addr,  key.validator_addr,  query)).data:(await (await initQueryClient(rootGetters)).queryDelegatorValidator( key.delegator_addr ,  key.validator_addr )).data
				
				commit('QUERY', { query: 'DelegatorValidator', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDelegatorValidator', payload: { options: { all }, params: {...key},query }})
				return getters['getDelegatorValidator']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryDelegatorValidator', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryHistoricalInfo({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryHistoricalInfo( key.height,  query)).data:(await (await initQueryClient(rootGetters)).queryHistoricalInfo( key.height )).data
				
				commit('QUERY', { query: 'HistoricalInfo', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryHistoricalInfo', payload: { options: { all }, params: {...key},query }})
				return getters['getHistoricalInfo']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryHistoricalInfo', 'API Node Unavailable. Could not perform query.'))
				return {}
			}
		},
		async QueryPool({ commit, rootGetters, getters }, { options: { subscribe = false , all = false}, params: {...key}, query=null }) {
			try {
				
				let value = query?(await (await initQueryClient(rootGetters)).queryPool( query)).data:(await (await initQueryClient(rootGetters)).queryPool()).data
				
				commit('QUERY', { query: 'Pool', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryPool', payload: { options: { all }, params: {...key},query }})
				return getters['getPool']( { params: {...key}, query}) ?? {}
			} catch (e) {
				console.error(new SpVuexError('QueryClient:QueryPool', 'API Node Unavailable. Could not perform query.'))
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
		
		async sendMsgDelegate({ rootGetters }, { value, fee, memo }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgDelegate(value)
				const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], {fee: { amount: fee, 
  gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgDelegate:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgDelegate:Send', 'Could not broadcast Tx.')
				}
			}
		},
		async sendMsgEditValidator({ rootGetters }, { value, fee, memo }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgEditValidator(value)
				const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], {fee: { amount: fee, 
  gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgEditValidator:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgEditValidator:Send', 'Could not broadcast Tx.')
				}
			}
		},
		async sendMsgCreateValidator({ rootGetters }, { value, fee, memo }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgCreateValidator(value)
				const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], {fee: { amount: fee, 
  gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgCreateValidator:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgCreateValidator:Send', 'Could not broadcast Tx.')
				}
			}
		},
		async sendMsgUndelegate({ rootGetters }, { value, fee, memo }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgUndelegate(value)
				const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], {fee: { amount: fee, 
  gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgUndelegate:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgUndelegate:Send', 'Could not broadcast Tx.')
				}
			}
		},
		async sendMsgBeginRedelegate({ rootGetters }, { value, fee, memo }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgBeginRedelegate(value)
				const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], {fee: { amount: fee, 
  gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgBeginRedelegate:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgBeginRedelegate:Send', 'Could not broadcast Tx.')
				}
			}
		},
		
		async MsgDelegate({ rootGetters }, { value }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgDelegate(value)
				return msg
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgDelegate:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgDelegate:Create', 'Could not create message.')
				}
			}
		},
		async MsgEditValidator({ rootGetters }, { value }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgEditValidator(value)
				return msg
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgEditValidator:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgEditValidator:Create', 'Could not create message.')
				}
			}
		},
		async MsgCreateValidator({ rootGetters }, { value }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgCreateValidator(value)
				return msg
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgCreateValidator:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgCreateValidator:Create', 'Could not create message.')
				}
			}
		},
		async MsgUndelegate({ rootGetters }, { value }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgUndelegate(value)
				return msg
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgUndelegate:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgUndelegate:Create', 'Could not create message.')
				}
			}
		},
		async MsgBeginRedelegate({ rootGetters }, { value }) {
			try {
				const msg = await (await initTxClient(rootGetters)).msgBeginRedelegate(value)
				return msg
			} catch (e) {
				if (e.toString()=='wallet is required') {
					throw new SpVuexError('TxClient:MsgBeginRedelegate:Init', 'Could not initialize signing client. Wallet is required.')
				}else{
					throw new SpVuexError('TxClient:MsgBeginRedelegate:Create', 'Could not create message.')
				}
			}
		},
		
	}
}
