import React from 'react'
import './Style.css'
import {useMetaMask} from "../hooks/useMetaMask";
import {useListen} from "../hooks/useListen";

type Props = {
    title: string
    caption: string | React.ReactNode
}

export const InitialView = () => {
    const {
        dispatch,
        state: {status, isMetaMaskInstalled, wallet},
    } = useMetaMask()
    const listen = useListen()

    // can be passed to an onclick handler
    const handleConnect = async () => {
        dispatch({type: 'loading'})
        const accounts = (await window.ethereum.request({
            method: 'eth_requestAccounts',
        })) as string[]

        if (accounts.length > 0) {
            const balance = (await window.ethereum!.request({
                method: 'eth_getBalance',
                params: [accounts[0], 'latest'],
            })) as string
            dispatch({type: 'connect', wallet: accounts[0], balance})

            // we can register an event listener for changes to the users wallet
            listen()
        }
    }

    return (
        <div className="wrapper">
            <div className="text-wrapper">
                <div className="heading-wrapper">
                    <h1>
                        <span className="heading">Decentralized Git</span>
                    </h1>
                    <p className="caption">The git storage, which can be [not]trusted</p>
                    <button onClick={handleConnect}>Connect MetaMask</button>
                </div>
            </div>
        </div>
    )
}
