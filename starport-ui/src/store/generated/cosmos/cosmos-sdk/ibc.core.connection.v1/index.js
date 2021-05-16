import { txClient, queryClient } from './module';
// @ts-ignore
import { SpVuexError } from '@starport/vuex';
import { ConnectionEnd } from "./module/types/ibc/core/connection/v1/connection";
import { IdentifiedConnection } from "./module/types/ibc/core/connection/v1/connection";
import { Counterparty } from "./module/types/ibc/core/connection/v1/connection";
import { ClientPaths } from "./module/types/ibc/core/connection/v1/connection";
import { ConnectionPaths } from "./module/types/ibc/core/connection/v1/connection";
import { Version } from "./module/types/ibc/core/connection/v1/connection";
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
        Connection: {},
        Connections: {},
        ClientConnections: {},
        ConnectionClientState: {},
        ConnectionConsensusState: {},
        _Structure: {
            ConnectionEnd: getStructure(ConnectionEnd.fromPartial({})),
            IdentifiedConnection: getStructure(IdentifiedConnection.fromPartial({})),
            Counterparty: getStructure(Counterparty.fromPartial({})),
            ClientPaths: getStructure(ClientPaths.fromPartial({})),
            ConnectionPaths: getStructure(ConnectionPaths.fromPartial({})),
            Version: getStructure(Version.fromPartial({})),
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
        getConnection: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.Connection[JSON.stringify(params)] ?? {};
        },
        getConnections: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.Connections[JSON.stringify(params)] ?? {};
        },
        getClientConnections: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.ClientConnections[JSON.stringify(params)] ?? {};
        },
        getConnectionClientState: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.ConnectionClientState[JSON.stringify(params)] ?? {};
        },
        getConnectionConsensusState: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.ConnectionConsensusState[JSON.stringify(params)] ?? {};
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
        async QueryConnection({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryConnection(key.connection_id, query)).data : (await (await initQueryClient(rootGetters)).queryConnection(key.connection_id)).data;
                commit('QUERY', { query: 'Connection', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryConnection', payload: { options: { all }, params: { ...key }, query } });
                return getters['getConnection']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryConnection', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryConnections({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryConnections(query)).data : (await (await initQueryClient(rootGetters)).queryConnections()).data;
                while (all && value.pagination && value.pagination.nextKey != null) {
                    let next_values = (await (await initQueryClient(rootGetters)).queryConnections({ ...query, 'pagination.key': value.pagination.nextKey })).data;
                    for (let prop of Object.keys(next_values)) {
                        if (Array.isArray(next_values[prop])) {
                            value[prop] = [...value[prop], ...next_values[prop]];
                        }
                        else {
                            value[prop] = next_values[prop];
                        }
                    }
                }
                commit('QUERY', { query: 'Connections', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryConnections', payload: { options: { all }, params: { ...key }, query } });
                return getters['getConnections']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryConnections', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryClientConnections({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryClientConnections(key.client_id, query)).data : (await (await initQueryClient(rootGetters)).queryClientConnections(key.client_id)).data;
                commit('QUERY', { query: 'ClientConnections', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryClientConnections', payload: { options: { all }, params: { ...key }, query } });
                return getters['getClientConnections']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryClientConnections', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryConnectionClientState({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryConnectionClientState(key.connection_id, query)).data : (await (await initQueryClient(rootGetters)).queryConnectionClientState(key.connection_id)).data;
                commit('QUERY', { query: 'ConnectionClientState', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryConnectionClientState', payload: { options: { all }, params: { ...key }, query } });
                return getters['getConnectionClientState']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryConnectionClientState', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryConnectionConsensusState({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryConnectionConsensusState(key.connection_id, key.revision_number, key.revision_height, query)).data : (await (await initQueryClient(rootGetters)).queryConnectionConsensusState(key.connection_id, key.revision_number, key.revision_height)).data;
                commit('QUERY', { query: 'ConnectionConsensusState', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryConnectionConsensusState', payload: { options: { all }, params: { ...key }, query } });
                return getters['getConnectionConsensusState']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryConnectionConsensusState', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async sendMsgConnectionOpenAck({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgConnectionOpenAck(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgConnectionOpenAck:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgConnectionOpenAck:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async sendMsgConnectionOpenConfirm({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgConnectionOpenConfirm(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgConnectionOpenConfirm:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgConnectionOpenConfirm:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async sendMsgConnectionOpenInit({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgConnectionOpenInit(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgConnectionOpenInit:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgConnectionOpenInit:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async sendMsgConnectionOpenTry({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgConnectionOpenTry(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgConnectionOpenTry:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgConnectionOpenTry:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async MsgConnectionOpenAck({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgConnectionOpenAck(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgConnectionOpenAck:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgConnectionOpenAck:Create', 'Could not create message.');
                }
            }
        },
        async MsgConnectionOpenConfirm({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgConnectionOpenConfirm(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgConnectionOpenConfirm:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgConnectionOpenConfirm:Create', 'Could not create message.');
                }
            }
        },
        async MsgConnectionOpenInit({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgConnectionOpenInit(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgConnectionOpenInit:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgConnectionOpenInit:Create', 'Could not create message.');
                }
            }
        },
        async MsgConnectionOpenTry({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgConnectionOpenTry(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgConnectionOpenTry:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgConnectionOpenTry:Create', 'Could not create message.');
                }
            }
        },
    }
};
