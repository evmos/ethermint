import HDWalletProvider from '@truffle/hdwallet-provider'
import Web3 from 'web3'
import BN = require('bn.js')
import {TransactionReceipt} from "web3-core"
import {Contract} from 'web3-eth-contract'
import {ethers} from "ethers"
const fs = require('fs')
const solc = require('solc')

const ctx: Array<((args:string[]) => Promise<void>)> = new Array<((args:string[]) => Promise<void>)>(999)
const ethJsonRpcUrl = 'http://127.0.0.1:8545'
const chainId = 9000

const USER1_ADDR="0xc6fe5d33615a1c52c08018c47e8bc53646a0e101"
const USER1_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"

const USER2_ADDR="0x963ebdf2e1f8db8707d05fc75bfeffba1b5bac17"
const USER2_MNEMONIC="maximum display century economy unlock van census kite error heart snow filter midnight usage egg venture cash kick motor survey drastic edge muffin visual"

ctx[1] = async function (args: string[]) {
    await handler(getMnemonicProvider(USER1_MNEMONIC) /* Use Wallet 1 */, async function(web3: Web3) {
        const sc = loadSmartContract(web3, 'WETH.sol', 'ERC20_Token_WETH')

        console.log('Deploying smart contract WETH.sol')
        const receipt1 = await deploySmartContract(web3, USER1_ADDR, USER1_MNEMONIC, sc)
        requireSuccessReceipt(receipt1)

        const contractAddr = receipt1?.contractAddress!

        console.log(`Sending ETH to new contract ${contractAddr}`)
        const receipt2 = await sendEther(web3, USER1_ADDR, contractAddr, 3)
        requireSuccessReceipt(receipt2)

        await handler(getMnemonicProvider(USER2_MNEMONIC) /* Use Wallet 2 */, async function(web3: Web3) {
            const contract = new web3.eth.Contract(sc.abi, contractAddr)

            console.log('Executing registerCollect()')
            const receipt3 = await contract.methods.registerCollect().send({
                from: USER2_ADDR,
                gas: 200000,
                gasPrice: web3.utils.toWei(new BN(1), "gwei")
            })
            requireSuccessReceipt(receipt3)

            console.log('Executing withdrawToRegisteredAddress()')
            const withdrawAmount = "2000000000000000" // 0.002 ETH
            const receipt4 = await contract.methods.withdrawToRegisteredAddress(withdrawAmount).send({
                from: USER2_ADDR,
                gas: 200000,
                gasPrice: web3.utils.toWei(new BN(1), "gwei")
            })
            requireSuccessReceipt(receipt4)
        })
    })
}

// Support

async function sendEther(web3: Web3, from: string | undefined, to: string | undefined, amount: number): Promise<TransactionReceipt> {
    return web3.eth.sendTransaction({
        from,
        to,
        value: web3.utils.toWei(new BN(amount), "ether"),
        chainId
    })
}

async function deploySmartContract(web3: Web3, from: string, mnemonic: string, smartContract: SmartContractInfo, argument?: any[]): Promise<TransactionReceipt | undefined> {
    const txObj = smartContract.contract.deploy({
        data: smartContract.bytecode,
        arguments: argument
    })
    const signedTx = await web3.eth.accounts.signTransaction(
        {
            data: txObj.encodeABI(),
            gas: 8000000,
        },
        getPrivateKeyFromMnemonic(mnemonic)
    )

    const receipt = await web3.eth.sendSignedTransaction(signedTx.rawTransaction as string)
    return receipt
}

async function main(args: string[]) {
    const testNo = parseInt(args[0])
    args = args.slice(1)
    await ctx[testNo](args)
}

async function handler(provider: HDWalletProvider, executor: (web3: Web3) => Promise<void>) {
    try {
        const web3 = new Web3(provider)
        await executor(web3)
    } catch (err) {
        console.error('err', err)
        provider.engine.stop()
        process.exit(1)
    } finally {
        provider.engine.stop()
    }
}

function getMnemonicProvider(mnemonic: string): HDWalletProvider {
    return new HDWalletProvider({
        mnemonic,
        providerOrUrl: ethJsonRpcUrl
    })
}

function getPrivateKeyFromMnemonic(mnemonic: string): string {
    return ethers.Wallet.fromMnemonic(mnemonic).privateKey
}

class SmartContractInfo {
    constructor(web3: Web3, content: string, contractClass: string) {
        const input = {
            language: 'Solidity',
            sources: {
                'Contract.sol': { content }
            },
            settings: {
                outputSelection: { '*': { '*': ['*'] } }
            }
        }

        const {contracts} = JSON.parse(
            solc.compile(JSON.stringify(input))
        );
        const smartContract = contracts[`Contract.sol`][contractClass]
        this.abi = smartContract.abi
        this.bytecode = smartContract.evm.bytecode.object
        this.contract = new web3.eth.Contract(this.abi)
    }

    contract: Contract;
    bytecode: any;
    abi: any;
}

function loadSmartContract(web3: Web3, fileName: string, contractClass: string): SmartContractInfo {
    const content = fs.readFileSync(`./tests/escan/contracts/${fileName}`, 'utf8')
    return new SmartContractInfo(web3, content, contractClass)
}

function requireSuccessReceipt(receipt: any) {
    if (!receipt) {
        console.error('No receipt')
        process.exit(1)
    }
    if (!receipt.status) {
        console.error('Not a success receipt', receipt)
        process.exit(1)
    }
}

main(process.argv.slice(2)).then()