import { txClient, queryClient } from './module';
// @ts-ignore
import { SpVuexError } from '@starport/vuex';
import { Params } from "./module/types/cosmos/bank/v1beta1/bank";
import { SendEnabled } from "./module/types/cosmos/bank/v1beta1/bank";
import { Input } from "./module/types/cosmos/bank/v1beta1/bank";
import { Output } from "./module/types/cosmos/bank/v1beta1/bank";
import { Supply } from "./module/types/cosmos/bank/v1beta1/bank";
import { DenomUnit } from "./module/types/cosmos/bank/v1beta1/bank";
import { Metadata } from "./module/types/cosmos/bank/v1beta1/bank";
import { Balance } from "./module/types/cosmos/bank/v1beta1/genesis";
async function initTxClient(vuexGetters) {
    return await txClient(vuexGetters['common/wallet/signer'], {
        addr: vuexGetters['common/env/apiTendermint']
    });
}
async function initQueryClient(vuexGetters) {
    return await queryClient({
        addr: vuexGetters['common/env/apiCosmos']
    });
}
function getStructure(template) {
    let structure = { fields: [] };
    for (const [key, value] of Object.entries(template)) {
        let field = {};
        field.name = key;
        field.type = typeof value;
        structure.fields.push(field);
    }
    return structure;
}
const getDefaultState = () => {
    return {
        Balance: {},
        AllBalances: {},
        TotalSupply: {},
        SupplyOf: {},
        Params: {},
        DenomMetadata: {},
        DenomsMetadata: {},
        _Structure: {
            Params: getStructure(Params.fromPartial({})),
            SendEnabled: getStructure(SendEnabled.fromPartial({})),
            Input: getStructure(Input.fromPartial({})),
            Output: getStructure(Output.fromPartial({})),
            Supply: getStructure(Supply.fromPartial({})),
            DenomUnit: getStructure(DenomUnit.fromPartial({})),
            Metadata: getStructure(Metadata.fromPartial({})),
            Balance: getStructure(Balance.fromPartial({})),
        },
        _Subscriptions: new Set(),
    };
};
// initial state
const state = getDefaultState();
export default {
    namespaced: true,
    state,
    mutations: {
        RESET_STATE(state) {
            Object.assign(state, getDefaultState());
        },
        QUERY(state, { query, key, value }) {
            state[query][JSON.stringify(key)] = value;
        },
        SUBSCRIBE(state, subscription) {
            state._Subscriptions.add(subscription);
        },
        UNSUBSCRIBE(state, subscription) {
            state._Subscriptions.delete(subscription);
        }
    },
    getters: {
        getBalance: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.Balance[JSON.stringify(params)] ?? {};
        },
        getAllBalances: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.AllBalances[JSON.stringify(params)] ?? {};
        },
        getTotalSupply: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.TotalSupply[JSON.stringify(params)] ?? {};
        },
        getSupplyOf: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.SupplyOf[JSON.stringify(params)] ?? {};
        },
        getParams: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.Params[JSON.stringify(params)] ?? {};
        },
        getDenomMetadata: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.DenomMetadata[JSON.stringify(params)] ?? {};
        },
        getDenomsMetadata: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.DenomsMetadata[JSON.stringify(params)] ?? {};
        },
        getTypeStructure: (state) => (type) => {
            return state._Structure[type].fields;
        }
    },
    actions: {
        init({ dispatch, rootGetters }) {
            console.log('init');
            if (rootGetters['common/env/client']) {
                rootGetters['common/env/client'].on('newblock', () => {
                    dispatch('StoreUpdate');
                });
            }
        },
        resetState({ commit }) {
            commit('RESET_STATE');
        },
        unsubscribe({ commit }, subscription) {
            commit('UNSUBSCRIBE', subscription);
        },
        async StoreUpdate({ state, dispatch }) {
            state._Subscriptions.forEach((subscription) => {
                dispatch(subscription.action, subscription.payload);
            });
        },
        async QueryBalance({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryBalance(key.address, key.denom, query)).data : (await (await initQueryClient(rootGetters)).queryBalance(key.address, key.denom)).data;
                commit('QUERY', { query: 'Balance', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryBalance', payload: { options: { all }, params: { ...key }, query } });
                return getters['getBalance']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryBalance', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryAllBalances({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryAllBalances(key.address, query)).data : (await (await initQueryClient(rootGetters)).queryAllBalances(key.address)).data;
                while (all && value.pagination && value.pagination.nextKey != null) {
                    let next_values = (await (await initQueryClient(rootGetters)).queryAllBalances(key.address, { ...query, 'pagination.key': value.pagination.nextKey })).data;
                    for (let prop of Object.keys(next_values)) {
                        if (Array.isArray(next_values[prop])) {
                            value[prop] = [...value[prop], ...next_values[prop]];
                        }
                        else {
                            value[prop] = next_values[prop];
                        }
                    }
                }
                commit('QUERY', { query: 'AllBalances', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryAllBalances', payload: { options: { all }, params: { ...key }, query } });
                return getters['getAllBalances']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryAllBalances', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryTotalSupply({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryTotalSupply(query)).data : (await (await initQueryClient(rootGetters)).queryTotalSupply()).data;
                commit('QUERY', { query: 'TotalSupply', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryTotalSupply', payload: { options: { all }, params: { ...key }, query } });
                return getters['getTotalSupply']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryTotalSupply', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QuerySupplyOf({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).querySupplyOf(key.denom, query)).data : (await (await initQueryClient(rootGetters)).querySupplyOf(key.denom)).data;
                commit('QUERY', { query: 'SupplyOf', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QuerySupplyOf', payload: { options: { all }, params: { ...key }, query } });
                return getters['getSupplyOf']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QuerySupplyOf', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryParams({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryParams(query)).data : (await (await initQueryClient(rootGetters)).queryParams()).data;
                commit('QUERY', { query: 'Params', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryParams', payload: { options: { all }, params: { ...key }, query } });
                return getters['getParams']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryParams', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryDenomMetadata({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryDenomMetadata(key.denom, query)).data : (await (await initQueryClient(rootGetters)).queryDenomMetadata(key.denom)).data;
                commit('QUERY', { query: 'DenomMetadata', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryDenomMetadata', payload: { options: { all }, params: { ...key }, query } });
                return getters['getDenomMetadata']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryDenomMetadata', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryDenomsMetadata({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryDenomsMetadata(query)).data : (await (await initQueryClient(rootGetters)).queryDenomsMetadata()).data;
                while (all && value.pagination && value.pagination.nextKey != null) {
                    let next_values = (await (await initQueryClient(rootGetters)).queryDenomsMetadata({ ...query, 'pagination.key': value.pagination.nextKey })).data;
                    for (let prop of Object.keys(next_values)) {
                        if (Array.isArray(next_values[prop])) {
                            value[prop] = [...value[prop], ...next_values[prop]];
                        }
                        else {
                            value[prop] = next_values[prop];
                        }
                    }
                }
                commit('QUERY', { query: 'DenomsMetadata', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryDenomsMetadata', payload: { options: { all }, params: { ...key }, query } });
                return getters['getDenomsMetadata']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryDenomsMetadata', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async sendMsgMultiSend({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgMultiSend(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgMultiSend:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgMultiSend:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async sendMsgSend({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgSend(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgSend:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgSend:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async MsgMultiSend({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgMultiSend(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgMultiSend:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgMultiSend:Create', 'Could not create message.');
                }
            }
        },
        async MsgSend({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgSend(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgSend:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgSend:Create', 'Could not create message.');
                }
            }
        },
    }
};
