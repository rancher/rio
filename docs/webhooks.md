# Webhook

By default, Rio will automatically poll the git repo and check if code has changed.
You can also configure a webhook to automatically push any events to Rio to trigger the build.

Note: Webhooks are currently only supported with Github.

1. Set up the GitHub webhook token.
```
$ rio secret add --github-webhook
Select namespace[default]: $(put the same namespace with your workload)
accessToken: $(github_accesstoken) # the token has to be able create webhook in your github repo.
Create workload and point to your repo.
```

2. Create a workload and point to your repo
```
rio run -p 80 --build-webhook-secret=githubtoken https://github.com/example/example
```

3. Go to your GitHub repo, it should have the webhook configured to point to one of the webhook services.

# Webhook for Riofile

Set up the webhook and private git clone secret for the git repository that contains Riofile:

```bash
$ rio up --build-clone-secret gitsecret --build-webhook-secret webhook https://github.com/example/example
```

For how to add the git and webhook secret, check [here](./continuous-deployment.md).

Note: `Riofile` and `Riofile-answers` in the root directory are automatically applied.
