import React from "react";
import {APIURL} from "../config";
import {useUser} from "../hooks/useUser";

type Props = {
    name: string
}

export const RepoRow: React.FC<Props> = ({name}) => {
    const {userToken} = useUser()

    const downloadAsZip = () => {
        const options = {
            headers: {
                Authorization: `Bearer ${userToken}`
            }
        };

        fetch(`${APIURL}/downloadRepo?` + new URLSearchParams({
            repo: name,
        }), options)
            .then(res => res.blob())
            .then(blob => {
                let file = window.URL.createObjectURL(blob);
                window.location.assign(file);
            });
    }

    return (
        <>
            <p>{name}</p>
            <button onClick={downloadAsZip}>Download</button>
        </>
    )
}
