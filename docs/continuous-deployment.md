# Continuous Deployment

Rio enables continuous deployment with GitHub by default.
This allows developers to streamline their focus on their git repository and worry less about their deploys.

The most versatile use case is in [this example](#pull-request-builds).


### Basic Example
Deploy a workload with Rio from a public GitHub repository that you have push access to:
`rio run -n cd-demo -p 8080 https://github.com/rancher/rio-demo`

Make a commit to the master branch of the repo. 
You should notice that within 15 seconds, Rio rebuilds your workload (`rio build-history`) and updates it to match the committed changes.


### Advanced Usage
#### Pull Request Builds
This feature uses a webhook to create a deployment when submitting a Pull Request to your tracked branch (master by default).
A new workload version will be staged in Rio, associated to the same app that was initially created. 
You can view the endpoint results directly from the PR by clicking "View deployment" in GitHub. 
If the pull request is merged, it will then update the app endpoint in Rio to point to this new version.

It only takes 2 steps:

1. [Configure Webhook](./webhooks.md) for your repository. For this example, you only need to set the webhook up.
2. `rio run -p 8080 -n example-cd --build-webhook-secret=githubtoken --build-pr --template https://github.com/example/example-repo`

NOTE: if your repository is private, you will also need to [create a credentials secret](#private-github-repo) and use the correct additional flags when running your workload.


#### Automatic Versioning
Notice the `--template` flag specified in the [Pull Request Builds](#pull-request-builds) scenario.
With this flag set, Rio will automatically configure versions for this workload when new commits are pushed to the GitHub repo.
As soon as the workload is ready, it will promote that version to have 100% of the app endpoint weight.
The only case where it won't automatically promote is when using the `--build-pr` flag as well and the build is from the PR branch.

If the `--template` flag is not set, then every subsequent build will overwrite the current version, including builds from the PR branch with `--build-pr` flag set.


#### Private Github Repo
You can do this with Git Basic Auth or SSH Auth:
- Git Basic Auth:
    1. Configure git basic auth credential secrets:
        ```bash
        $ rio secret create --git-basic-auth
        Select namespace[default]: $(put the same namespace with your workload)
        git url[]: $(for example: https://github.com)
        username[]: $(your GH username)
        password[******]: $(your GH password)
        ```
    2. Create a workload pointing to your repo using standard git checkout. For example:
        `rio run -p 8080 https://github.com/example/example-private-repo`
- SSH Auth:
    1. Configure git sshkey auth credential secrets. This should use a key that does not have a password associated to it:
        ```bash
        $ rio secret create --git-sshkey-auth
        Select namespace[default]: $(put the same namespace with your workload)
        git url[]: $(put your github url. Leave out http/https. Example: github.com)
        ssh_key_path[]: $(type the path to your ssh private key)
        ```
    2. Create workload pointing to your repo using ssh checkout. For example:
        `rio run --build-clone-secret gitcredential-ssh -p 8080 git@github.com:example/example.git`


#### Private Docker Registry
1. Configure the Docker credential secret.
```bash
$ rio secret create --docker
Select namespace[default]: $(put the same namespace with your workload)
Registry URL[https://index.docker.io/v1/]: $(found with "docker info | grep Registry")
username[]: $(your docker username)
password[******]: $(your docker password)
```
2. Create a workload pointing to your image. For example:
`rio run --image-pull-secrets dockerconfig -p 8080 imageorg/imagename:version`


### Useful Options
There are many options available for use when running workloads in Rio. These are just a few that are useful for CD:

| Option | Type | Description |
|------|----| -------------|
| `--build-branch` | string | Build repository branch (default: "master") | 
| `--build-dockerfile` | string | Set Dockerfile name, defaults to Dockerfile |
| `--build-context` | string | Set build context, defaults to . |
| `--build-webhook-secret` | string | Set GitHub webhook secret name |
| `--build-clone-secret` | string | Set git clone secret name |
| `--build-image-name` | string | Specify custom image name to push |
| `--build-pr` | boolean | Enable pull request builds |
| `--build-timeout` | string | Timeout for build, default to 10m (ms|s|m|h) |
| `--image-pull-secrets` | string | Specify image pull secrets |
| `--template` | boolean | If true new version is created per git commit. If false update in-place |
