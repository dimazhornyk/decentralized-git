import React, {useEffect, useState} from 'react'
import './Style.css'
import {useMetaMask} from "../hooks/useMetaMask"
import {APIURL} from "../config"

type Props = {
    name: string
}

const RepoRow: React.FC<Props> = ({name}) => {
    return (
        <p>{name}</p>
    )
}

export const ReposView = () => {
    const {dispatch, state: {status, isMetaMaskInstalled, wallet}} = useMetaMask()
    const [repos, setRepos] = useState([]);

    useEffect(() => {
        fetch(`${APIURL}/getRepos?` + new URLSearchParams({
            wallet: wallet!,
        }))
            .then((res) => res.json())
            .then((data) => {
                setRepos(data)
            })
            .catch((err) => {
                console.log(err.message);
            });
    }, []);

    return (
        <div>
            {repos.map((repo, i) => <RepoRow name={repo} key={i}/>)}
        </div>
    )
}
