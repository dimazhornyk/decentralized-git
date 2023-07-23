import {useState, useEffect} from 'react';
import {APIURL} from "../config";
import {useMetaMask} from "./useMetaMask";

export function useUser() {
    const {dispatch, state: {status, isMetaMaskInstalled, wallet}} = useMetaMask()

    const [userToken, setUserToken] = useState(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        async function getUserDetails() {
            const token = await getAuthenticatedUser(wallet!);
            if (token === undefined) {
                console.log("error getting user auth")
                return
            }

            setUserToken(token);
            setIsLoading(false);
        }

        getUserDetails();
    }, []);

    return {userToken, isLoading};
}

const getAuthenticatedUser = async (wallet: string) => {
    const nonceRes = await fetch(`${APIURL}/nonce`)
    const nonceData = await nonceRes.json()

    let resp = await siweSign(wallet, nonceData.nonce)
    if (resp === undefined) {
        throw "can't get a signature"
    }

    let res = await fetch(`${APIURL}/login`, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({message: resp!.message, signature: resp!.signature})
    })
    if (res.status !== 200) {
        res = await fetch(`${APIURL}/register`, {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({message: resp!.message, signature: resp!.signature})
        })
        const data = await res.json()
        window.alert(`Congratulations, you have been registered.\nHere is your GitHub actions token: ${data.action_token} and encryption key ${data.encryption_key}`)
        console.log(`Congratulations, you have been registered.\nHere is your GitHub actions token: ${data.action_token} and encryption key ${data.encryption_key}`)
    }

    res = await fetch(`${APIURL}/login`, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({message: resp!.message, signature: resp!.signature})
    })

    const data = await res.json()
    return data.token
}

const siweSign = async (wallet: string, nonce: string) => {
    try {
        console.log(wallet)
        const msgText = `${window.location.host} wants you to sign in with your Ethereum account:\n${wallet}\n\nThis is a test statement.\n\nURI: https://${window.location.host}\nVersion: 1\nChain ID: 1\nNonce: ${nonce}\nIssued At: 2021-09-30T16:25:24.000Z`

        // const msg = `0x${Buffer.from(msgText, 'utf8').toString('hex')}`;
        const sign = await window.ethereum!.request({
            method: 'personal_sign',
            params: [msgText, wallet],
        });

        return {message: msgText, signature: sign}
    } catch (err) {
        console.error(err);
    }
};
