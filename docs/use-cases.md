## Using Rio to rollout your service mutliple times a week (Part-time Devops)

### Image-only ( use Rio as a way to easily rollout service level updates)

- Command to run: `rio run $image`

- Action to update: `something`

- How to use: Use Rio to run your image and then use rio router with your CNAME to shift traffic to the new image version. You can also use the URL in `rio ps` to see what your service is like before rolling it out to the production URL in Rio router.



### Build from a git source repository, update in-place (polling or webhook)

- Your commnd" `rio run $repo` (with options to tell rio to ovewrite this with latest)

- Action to update: `rio something`

- How to use: Use  Rio to 


### Build from source code, use template with polling or webhook, update on every commit

### Build from git source repository, no updates to (production) service, manual stage for every new commit. 


### Build from source code, update production service with every tagged release

### Build from source code, update production service with every tagged releaes (deploy service); potentially auto-rollout from every tag release