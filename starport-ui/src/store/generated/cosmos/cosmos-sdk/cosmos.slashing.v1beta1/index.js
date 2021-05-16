import { txClient, queryClient } from './module';
// @ts-ignore
import { SpVuexError } from '@starport/vuex';
import { ValidatorSigningInfo } from "./module/types/cosmos/slashing/v1beta1/slashing";
import { Params } from "./module/types/cosmos/slashing/v1beta1/slashing";
import { SigningInfo } from "./module/types/cosmos/slashing/v1beta1/genesis";
import { ValidatorMissedBlocks } from "./module/types/cosmos/slashing/v1beta1/genesis";
import { MissedBlock } from "./module/types/cosmos/slashing/v1beta1/genesis";
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
        Params: {},
        SigningInfo: {},
        SigningInfos: {},
        _Structure: {
            ValidatorSigningInfo: getStructure(ValidatorSigningInfo.fromPartial({})),
            Params: getStructure(Params.fromPartial({})),
            SigningInfo: getStructure(SigningInfo.fromPartial({})),
            ValidatorMissedBlocks: getStructure(ValidatorMissedBlocks.fromPartial({})),
            MissedBlock: getStructure(MissedBlock.fromPartial({})),
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
        getParams: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.Params[JSON.stringify(params)] ?? {};
        },
        getSigningInfo: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.SigningInfo[JSON.stringify(params)] ?? {};
        },
        getSigningInfos: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.SigningInfos[JSON.stringify(params)] ?? {};
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
        async QuerySigningInfo({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).querySigningInfo(key.cons_address, query)).data : (await (await initQueryClient(rootGetters)).querySigningInfo(key.cons_address)).data;
                commit('QUERY', { query: 'SigningInfo', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QuerySigningInfo', payload: { options: { all }, params: { ...key }, query } });
                return getters['getSigningInfo']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QuerySigningInfo', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QuerySigningInfos({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).querySigningInfos(query)).data : (await (await initQueryClient(rootGetters)).querySigningInfos()).data;
                while (all && value.pagination && value.pagination.nextKey != null) {
                    let next_values = (await (await initQueryClient(rootGetters)).querySigningInfos({ ...query, 'pagination.key': value.pagination.nextKey })).data;
                    for (let prop of Object.keys(next_values)) {
                        if (Array.isArray(next_values[prop])) {
                            value[prop] = [...value[prop], ...next_values[prop]];
                        }
                        else {
                            value[prop] = next_values[prop];
                        }
                    }
                }
                commit('QUERY', { query: 'SigningInfos', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QuerySigningInfos', payload: { options: { all }, params: { ...key }, query } });
                return getters['getSigningInfos']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QuerySigningInfos', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async sendMsgUnjail({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgUnjail(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgUnjail:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgUnjail:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async MsgUnjail({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgUnjail(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgUnjail:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgUnjail:Create', 'Could not create message.');
                }
            }
        },
    }
};
