import { txClient, queryClient } from './module';
// @ts-ignore
import { SpVuexError } from '@starport/vuex';
import { Params } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { ValidatorHistoricalRewards } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { ValidatorCurrentRewards } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { ValidatorAccumulatedCommission } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { ValidatorOutstandingRewards } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { ValidatorSlashEvent } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { ValidatorSlashEvents } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { FeePool } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { CommunityPoolSpendProposal } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { DelegatorStartingInfo } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { DelegationDelegatorReward } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { CommunityPoolSpendProposalWithDeposit } from "./module/types/cosmos/distribution/v1beta1/distribution";
import { DelegatorWithdrawInfo } from "./module/types/cosmos/distribution/v1beta1/genesis";
import { ValidatorOutstandingRewardsRecord } from "./module/types/cosmos/distribution/v1beta1/genesis";
import { ValidatorAccumulatedCommissionRecord } from "./module/types/cosmos/distribution/v1beta1/genesis";
import { ValidatorHistoricalRewardsRecord } from "./module/types/cosmos/distribution/v1beta1/genesis";
import { ValidatorCurrentRewardsRecord } from "./module/types/cosmos/distribution/v1beta1/genesis";
import { DelegatorStartingInfoRecord } from "./module/types/cosmos/distribution/v1beta1/genesis";
import { ValidatorSlashEventRecord } from "./module/types/cosmos/distribution/v1beta1/genesis";
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
        ValidatorOutstandingRewards: {},
        ValidatorCommission: {},
        ValidatorSlashes: {},
        DelegationRewards: {},
        DelegationTotalRewards: {},
        DelegatorValidators: {},
        DelegatorWithdrawAddress: {},
        CommunityPool: {},
        _Structure: {
            Params: getStructure(Params.fromPartial({})),
            ValidatorHistoricalRewards: getStructure(ValidatorHistoricalRewards.fromPartial({})),
            ValidatorCurrentRewards: getStructure(ValidatorCurrentRewards.fromPartial({})),
            ValidatorAccumulatedCommission: getStructure(ValidatorAccumulatedCommission.fromPartial({})),
            ValidatorOutstandingRewards: getStructure(ValidatorOutstandingRewards.fromPartial({})),
            ValidatorSlashEvent: getStructure(ValidatorSlashEvent.fromPartial({})),
            ValidatorSlashEvents: getStructure(ValidatorSlashEvents.fromPartial({})),
            FeePool: getStructure(FeePool.fromPartial({})),
            CommunityPoolSpendProposal: getStructure(CommunityPoolSpendProposal.fromPartial({})),
            DelegatorStartingInfo: getStructure(DelegatorStartingInfo.fromPartial({})),
            DelegationDelegatorReward: getStructure(DelegationDelegatorReward.fromPartial({})),
            CommunityPoolSpendProposalWithDeposit: getStructure(CommunityPoolSpendProposalWithDeposit.fromPartial({})),
            DelegatorWithdrawInfo: getStructure(DelegatorWithdrawInfo.fromPartial({})),
            ValidatorOutstandingRewardsRecord: getStructure(ValidatorOutstandingRewardsRecord.fromPartial({})),
            ValidatorAccumulatedCommissionRecord: getStructure(ValidatorAccumulatedCommissionRecord.fromPartial({})),
            ValidatorHistoricalRewardsRecord: getStructure(ValidatorHistoricalRewardsRecord.fromPartial({})),
            ValidatorCurrentRewardsRecord: getStructure(ValidatorCurrentRewardsRecord.fromPartial({})),
            DelegatorStartingInfoRecord: getStructure(DelegatorStartingInfoRecord.fromPartial({})),
            ValidatorSlashEventRecord: getStructure(ValidatorSlashEventRecord.fromPartial({})),
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
        getValidatorOutstandingRewards: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.ValidatorOutstandingRewards[JSON.stringify(params)] ?? {};
        },
        getValidatorCommission: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.ValidatorCommission[JSON.stringify(params)] ?? {};
        },
        getValidatorSlashes: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.ValidatorSlashes[JSON.stringify(params)] ?? {};
        },
        getDelegationRewards: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.DelegationRewards[JSON.stringify(params)] ?? {};
        },
        getDelegationTotalRewards: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.DelegationTotalRewards[JSON.stringify(params)] ?? {};
        },
        getDelegatorValidators: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.DelegatorValidators[JSON.stringify(params)] ?? {};
        },
        getDelegatorWithdrawAddress: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.DelegatorWithdrawAddress[JSON.stringify(params)] ?? {};
        },
        getCommunityPool: (state) => (params = {}) => {
            if (!params.query) {
                params.query = null;
            }
            return state.CommunityPool[JSON.stringify(params)] ?? {};
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
        async QueryValidatorOutstandingRewards({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryValidatorOutstandingRewards(key.validator_address, query)).data : (await (await initQueryClient(rootGetters)).queryValidatorOutstandingRewards(key.validator_address)).data;
                commit('QUERY', { query: 'ValidatorOutstandingRewards', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryValidatorOutstandingRewards', payload: { options: { all }, params: { ...key }, query } });
                return getters['getValidatorOutstandingRewards']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryValidatorOutstandingRewards', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryValidatorCommission({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryValidatorCommission(key.validator_address, query)).data : (await (await initQueryClient(rootGetters)).queryValidatorCommission(key.validator_address)).data;
                commit('QUERY', { query: 'ValidatorCommission', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryValidatorCommission', payload: { options: { all }, params: { ...key }, query } });
                return getters['getValidatorCommission']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryValidatorCommission', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryValidatorSlashes({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryValidatorSlashes(key.validator_address, query)).data : (await (await initQueryClient(rootGetters)).queryValidatorSlashes(key.validator_address)).data;
                while (all && value.pagination && value.pagination.nextKey != null) {
                    let next_values = (await (await initQueryClient(rootGetters)).queryValidatorSlashes(key.validator_address, { ...query, 'pagination.key': value.pagination.nextKey })).data;
                    for (let prop of Object.keys(next_values)) {
                        if (Array.isArray(next_values[prop])) {
                            value[prop] = [...value[prop], ...next_values[prop]];
                        }
                        else {
                            value[prop] = next_values[prop];
                        }
                    }
                }
                commit('QUERY', { query: 'ValidatorSlashes', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryValidatorSlashes', payload: { options: { all }, params: { ...key }, query } });
                return getters['getValidatorSlashes']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryValidatorSlashes', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryDelegationRewards({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryDelegationRewards(key.delegator_address, key.validator_address, query)).data : (await (await initQueryClient(rootGetters)).queryDelegationRewards(key.delegator_address, key.validator_address)).data;
                commit('QUERY', { query: 'DelegationRewards', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryDelegationRewards', payload: { options: { all }, params: { ...key }, query } });
                return getters['getDelegationRewards']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryDelegationRewards', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryDelegationTotalRewards({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryDelegationTotalRewards(key.delegator_address, query)).data : (await (await initQueryClient(rootGetters)).queryDelegationTotalRewards(key.delegator_address)).data;
                commit('QUERY', { query: 'DelegationTotalRewards', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryDelegationTotalRewards', payload: { options: { all }, params: { ...key }, query } });
                return getters['getDelegationTotalRewards']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryDelegationTotalRewards', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryDelegatorValidators({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryDelegatorValidators(key.delegator_address, query)).data : (await (await initQueryClient(rootGetters)).queryDelegatorValidators(key.delegator_address)).data;
                commit('QUERY', { query: 'DelegatorValidators', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryDelegatorValidators', payload: { options: { all }, params: { ...key }, query } });
                return getters['getDelegatorValidators']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryDelegatorValidators', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryDelegatorWithdrawAddress({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryDelegatorWithdrawAddress(key.delegator_address, query)).data : (await (await initQueryClient(rootGetters)).queryDelegatorWithdrawAddress(key.delegator_address)).data;
                commit('QUERY', { query: 'DelegatorWithdrawAddress', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryDelegatorWithdrawAddress', payload: { options: { all }, params: { ...key }, query } });
                return getters['getDelegatorWithdrawAddress']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryDelegatorWithdrawAddress', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async QueryCommunityPool({ commit, rootGetters, getters }, { options: { subscribe = false, all = false }, params: { ...key }, query = null }) {
            try {
                let value = query ? (await (await initQueryClient(rootGetters)).queryCommunityPool(query)).data : (await (await initQueryClient(rootGetters)).queryCommunityPool()).data;
                commit('QUERY', { query: 'CommunityPool', key: { params: { ...key }, query }, value });
                if (subscribe)
                    commit('SUBSCRIBE', { action: 'QueryCommunityPool', payload: { options: { all }, params: { ...key }, query } });
                return getters['getCommunityPool']({ params: { ...key }, query }) ?? {};
            }
            catch (e) {
                console.error(new SpVuexError('QueryClient:QueryCommunityPool', 'API Node Unavailable. Could not perform query.'));
                return {};
            }
        },
        async sendMsgFundCommunityPool({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgFundCommunityPool(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgFundCommunityPool:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgFundCommunityPool:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async sendMsgWithdrawValidatorCommission({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgWithdrawValidatorCommission(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgWithdrawValidatorCommission:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgWithdrawValidatorCommission:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async sendMsgWithdrawDelegatorReward({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgWithdrawDelegatorReward(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgWithdrawDelegatorReward:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgWithdrawDelegatorReward:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async sendMsgSetWithdrawAddress({ rootGetters }, { value, fee, memo }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgSetWithdrawAddress(value);
                const result = await (await initTxClient(rootGetters)).signAndBroadcast([msg], { fee: { amount: fee,
                        gas: "200000" }, memo });
                return result;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgSetWithdrawAddress:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgSetWithdrawAddress:Send', 'Could not broadcast Tx.');
                }
            }
        },
        async MsgFundCommunityPool({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgFundCommunityPool(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgFundCommunityPool:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgFundCommunityPool:Create', 'Could not create message.');
                }
            }
        },
        async MsgWithdrawValidatorCommission({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgWithdrawValidatorCommission(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgWithdrawValidatorCommission:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgWithdrawValidatorCommission:Create', 'Could not create message.');
                }
            }
        },
        async MsgWithdrawDelegatorReward({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgWithdrawDelegatorReward(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgWithdrawDelegatorReward:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgWithdrawDelegatorReward:Create', 'Could not create message.');
                }
            }
        },
        async MsgSetWithdrawAddress({ rootGetters }, { value }) {
            try {
                const msg = await (await initTxClient(rootGetters)).msgSetWithdrawAddress(value);
                return msg;
            }
            catch (e) {
                if (e.toString() == 'wallet is required') {
                    throw new SpVuexError('TxClient:MsgSetWithdrawAddress:Init', 'Could not initialize signing client. Wallet is required.');
                }
                else {
                    throw new SpVuexError('TxClient:MsgSetWithdrawAddress:Create', 'Could not create message.');
                }
            }
        },
    }
};
