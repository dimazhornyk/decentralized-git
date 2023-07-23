import React from "react";
import {APIURL} from "../config";
import {useUser} from "../hooks/useUser";

type Props = {
    name: string,
    token: string
}

export const RepoRow: React.FC<Props> = ({name, token}) => {
    const downloadAsZip = () => {
        const options = {
            headers: {
                Authorization: `Bearer ${token}`
            }
        };

        fetch(`${APIURL}/api/downloadRepo?` + new URLSearchParams({
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
