declare const _default: {
    namespaced: boolean;
    state: {
        _Structure: {
            TxResponse: {
                fields: any[];
            };
            ABCIMessageLog: {
                fields: any[];
            };
            StringEvent: {
                fields: any[];
            };
            Attribute: {
                fields: any[];
            };
            GasInfo: {
                fields: any[];
            };
            Result: {
                fields: any[];
            };
            SimulationResponse: {
                fields: any[];
            };
            MsgData: {
                fields: any[];
            };
            TxMsgData: {
                fields: any[];
            };
            SearchTxsResult: {
                fields: any[];
            };
        };
        _Subscriptions: Set<unknown>;
    };
    mutations: {
        RESET_STATE(state: any): void;
        QUERY(state: any, { query, key, value }: {
            query: any;
            key: any;
            value: any;
        }): void;
        SUBSCRIBE(state: any, subscription: any): void;
        UNSUBSCRIBE(state: any, subscription: any): void;
    };
    getters: {
        getTypeStructure: (state: any) => (type: any) => any;
    };
    actions: {
        init({ dispatch, rootGetters }: {
            dispatch: any;
            rootGetters: any;
        }): void;
        resetState({ commit }: {
            commit: any;
        }): void;
        unsubscribe({ commit }: {
            commit: any;
        }, subscription: any): void;
        StoreUpdate({ state, dispatch }: {
            state: any;
            dispatch: any;
        }): Promise<void>;
    };
};
export default _default;
