## Purpose
Thanks for stopping by. I use this doc to learn and note down the deployment process. It tries to contain all of my thought processes and steps for the deployment of this CLI tool.


# Some deployment notes
- PAT (personal access token) seems to have TTL, so if brew release fails, check if the token is still valid. If not, generate a new one and update the secrets in GitHub.
- Brew release may fail at some point. But it is probably not a good idea to fix it since Apple is greedy and wants us to pay $99 per month to allow us to distribute the binary via Homebrew Cask.
- The best way is to use github releases and let users download the tar.gz file from the releases page.
