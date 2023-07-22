import React, {useEffect, useState} from 'react'
import './Style.css'
import {APIURL} from "../config"
import {RepoRow} from "./RepoRow";
import {useUser} from "../hooks/useUser";


export const ReposView = () => {
    const [repos, setRepos] = useState([]);
    const { userToken } = useUser();

    useEffect(() => {
        fetch(`${APIURL}/getRepos`, {
            headers: {
                Authorization: `Bearer ${userToken}`
            }
        })
            .then((res) => res.json())
            .then((data) => {
                setRepos(data.repos)
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
