import {useState, useEffect} from 'react';
import {APIURL} from "../config";
import {useMetaMask} from "./useMetaMask";

export function useUser() {
    const {dispatch, state: {status, isMetaMaskInstalled, wallet}} = useMetaMask()

    const [userToken, setUserToken] = useState(null);

    useEffect(() => {
        async function getUserDetails() {
            const token = await getAuthenticatedUser(wallet!);
            if (token === undefined) {
                console.log("error getting user auth")
                return
            }

            setUserToken(token);
        }

        getUserDetails();
    }, []);

    return {userToken};
}

const getAuthenticatedUser = async (wallet: string) => {
    let resp = await siweSign(wallet)
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

const siweSign = async (wallet: string) => {
    try {
        const msg = `0x${Buffer.from("Some msg!", 'utf8').toString('hex')}`;
        const sign = await window.ethereum!.request({
            method: 'personal_sign',
            params: [msg, wallet],
        });

        return {message: msg, signature: sign}
    } catch (err) {
        console.error(err);
    }
};
