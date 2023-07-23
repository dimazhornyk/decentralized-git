import React, {useEffect, useState} from 'react'
import './Style.css'
import {APIURL} from "../config"
import {RepoRow} from "./RepoRow";
import {useUser} from "../hooks/useUser";


export const ReposView = () => {
    const [repos, setRepos] = useState([]);
    const {userToken, isLoading} = useUser();

    useEffect(() => {
        if (isLoading) {
            return
        }

        fetch(`${APIURL}/api/getRepos`, {
            headers: {
                Authorization: `Bearer ${userToken}`
            }
        })
            .then((res) => res.json())
            .then((data) => {
                console.log(data.repos)
                setRepos(data.repos)
            })
            .catch((err) => {
                console.log(err.message);
            });
    }, [isLoading]);

    return (
        <div style={{display: "flex", flexDirection: "column", alignItems: "center"}}>
            <h2>Repositories:</h2>
            {repos.map((repo, i) => <RepoRow name={repo} token={userToken!} key={i}/>)}
        </div>
    )
}
