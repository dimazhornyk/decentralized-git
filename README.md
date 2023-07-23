# Git storage on Filecoin

### This service is being accessed both by the GitHub action that uploads the zipped files from every commit and by the client-side application
## Backend
- Uses SIWE for user sign-in
- Based on that data generates a session tokens that are used in further communication
- Uses custom diff-storing format that satisfies the git-related use cases the most and the space consumption
- Generates encryption key and GitHub Action key upon user registration
- Uses AES to encrypt the data that's being stored
#### Depends on w3s for interacting with a Filecoin
Is called by this Action https://github.com/dimazhornyk/decentralized-git-uploader-action

## Frontend
- Uses React with TypeScrypt
- Uses Metamask SDK and generates a signature to prove the ownership of the account

### Improvements:
- Add endpoints for a complete files browsing solution, including commit-based historical views
- Move to onchain access-control e.g. ZK-Groups for access to the certain repositories
- Handle all the cases of concurrency and usage of the same disk space/memory
