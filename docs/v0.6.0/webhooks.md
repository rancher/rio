# Webhook

By default, rio will automatically pull git repo and check if repo code has changed.
You can also configure a webhook to automatically push any events to Rio to trigger the build.

1. Set up Github webhook token.
```
$ rio secret add --github-webhook
Select namespace[default]: $(put the same namespace with your workload)
accessToken: $(github_accesstoken) # the token has to be able create webhook in your github repo.
Create workload and point to your repo.
```

2. Create workload and point to your repo
```
rio run -p 80 --build-webhook-secret=githubtoken https://github.com/example/example
```

3. Go to your Github repo, it should have webhook configured to point to one of our webhook service.

