## Purpose
Thanks for stopping by. I use this doc to learn and note down the deployment process. It tries to contain all of my thought processes and steps for the deployment of this CLI tool.

## Brew public release flow
With the help of GPT, the goreleasr apparently uses the following flow to update both brew cask and brew tap repos.
1. A tag build triggers .github/workflows/release.yml
2. The workflow starts GoReleaser and passes secrets.GH_PAT as GITHUB_TOKEN
3. GoReleaser reads .goreleaser.yaml, sees my homebrew cask repository and uses the runtime token in the previous step to write to the tap repo. 
4. The resulting commits in goreleaser/homebrew-tap show up as comign from goreleaserbot. The tap repo’s commit history is full of automated “Brew cask update …” commits by goreleaserbot.

As of April 2026, brews command in .goreleaser.yaml is deprecated, so I have to use homebrew_casks instead, although homebrew casks seems to be for GUI apps, but I guess it works for CLI tools as well.

# Potential troubleshoot steps
- PAT (personal access token) seems to have TTL, so if brew release fails, check if the token is still valid. If not, generate a new one and update the secrets in GitHub.