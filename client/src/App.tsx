import {BrowserRouter, Route, Routes, Navigate} from 'react-router-dom';
import {MetaMaskProvider, useMetaMask} from './hooks/useMetaMask'
import {SdkLayout} from './components/SdkProvider'
import {InitialView} from "./components/Initial";
import {ReposView} from "./components/Repos";

export default function App() {
    return (
        <MetaMaskProvider>
            <SdkLayout>
                {/*<BrowserRouter>*/}
                {/*    <Routes>*/}
                {/*        <Route path="/" element={<Main/>}/>*/}
                <Main/>
                {/*</Routes>*/}
                {/*</BrowserRouter>*/}
            </SdkLayout>
        </MetaMaskProvider>
    )
}

function Main() {
    const {
        dispatch,
        state: {status, isMetaMaskInstalled, wallet},
    } = useMetaMask()

    console.log('hello')
    // we can use this to conditionally render the UI
    const showInstallMetaMask = status !== 'pageNotLoaded' && !isMetaMaskInstalled

    // we can use this to conditionally render the UI
    const showConnectButton = status !== 'pageNotLoaded' && isMetaMaskInstalled && !wallet

    // we can use this to conditionally render the UI
    const isConnected = status !== 'pageNotLoaded' && typeof wallet === 'string'

    // can be passed to an onclick handler
    const handleDisconnect = () => {
        dispatch({type: 'disconnect'})
    }

    return (
        <>
            {showConnectButton || !isConnected ? <InitialView/> : <ReposView/>}
        </>
    )
}
