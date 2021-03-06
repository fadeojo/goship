[![Build Status](https://travis-ci.org/gengo/goship.svg?branch=master)](https://travis-ci.org/gengo/goship)

# GoShip

A simple tool for deploying code to servers.

![pirate gopher](https://cloud.githubusercontent.com/assets/3772659/8693461/3c5f74a8-2b12-11e5-9a27-ff4421589df6.png)

# What it does:

GoShip SSHes into the machines that you list in ETCD and gets the latest revision from the specified git repository. It then compares that to the latest revision on GitHub and, if they differ, shows a link to the diff as well as a Deploy button. You can then deploy by clicking the button, and will show you the output of the deployment command, as well as save the output, diff, and whether the command succeeded.

![GoShip Index Page Screenshot](https://cloud.githubusercontent.com/assets/3772659/8693471/55ec2592-2b12-11e5-965f-8e572309c945.png)

# Installation
1. Install [go](http://golang.org) development environment
2. run `go get github.com/gengo/goship`
3. run `go build github.com/gengo/goship`

# Usage

1. Export your GitHub API token:
   
   ```shell
   export GITHUB_API_TOKEN="your-organization-github-token-here"
   ```
2. Github Omniauth Integration:
   
   Users who are collaborator on a repo can 'see' that repo in Goship.
   You must create a developer application to use omniauth.
   If you do NOT add the appropriate env keys below AUTH will be OFF I.E. Please be careful and check the logs.
   Please  note the "Authorization callback URL" should match your site i.e. `http://<your-url-and-port>/auth/github/callback`.

   ```shell
   export GITHUB_RANDOM_HASH_KEY="some-random-hash-here";
   export GITHUB_OMNI_AUTH_ID="github-application-id";
   export GITHUB_OMNI_AUTH_KEY="github-application-key";
   export GITHUB_CALLBACK_URL="http://<your-url-and-port>";  // must match that given to Github! Would be 127.0.0.1:port for testing
   ```
   
   If authentication is 'turned on', organization 'team' members who are collaborators and exclusively on a 'pull' only team will be able to see a repo, however the deploy button will be diasbled for them.
   
3. Create an etcd server
   1. Follow the instructions in the [etcd](https://github.com/coreos/etcd) README
   2. Write configurations in YAML.
   3. Store the configuration into etcd with `goshipcfg` in [tools](#tools)
      * You can also directly access to etcd entries for small amount of change.
        * etcd exposes a single set of APIs. [[reference](https://coreos.com/etcd/docs/latest/api.html)].
          There are many tools to call the APIs, e.g. [etcdctl](https://github.com/coreos/etcdctl/).


# Example
Configure a file `config.yaml`.
   ```yaml
   deploy_user: YOUR_SSH_USER_ON_SERVER
   projects:
   - name: my-project
     repo_name: my-project
     repo_owner: github-user-or-org
     envs:
     - name: staging
       deploy: "/tmp/deploy -p=my-project -e=staging"
       repo_path: "PATH_TO_REPOSITORY/.git"
       hosts:
       - my-staging-server.example.com
       branch: master
   ```

Run this command.
   ```shell
   goshipcfg -store -logtostderr < config.yaml
   ```

A quick explaination of keys used in this sample structure:
* **deploy_user:** This is your SSH user on the application server that Goship SSH user will have password-less auth to
* **repo_name:** Name of your application project repository
* **repo_owner:** Name of your Github user, or your Github org which owns the repo
* **deploy:** This is your deploy command with necessary arguments. A sample script is included(tools/deploy)
* **repo_path:** Path to your application code repository on the application server
* **hosts:** An array of FQDN of the host(s), where Goship will deploy the code
* **branch:** Application code branch to deploy
* **comment:** Any comments/notes

# Commandline Flags

```
 -b [bind address]                   Address to bind (default localhost:8000)
 -d [data path]                      Path to data directory (default ./data/)
 -e [etcd location]                  Full URL to ETCD Server (default http://127.0.0.1:4001)
 -k [id_rsa key]                     Path to private SSH key for connecting to Github (default id_rsa)
 -s [static files]                   Path to directory for static files (default ./static/)
 -request-log [request log path]     Destination of request log (default '-', which is stdout)
```

Run `goship -help` for more flags.

# Chat Notifications
To notify a chat room when the Deploy button is pushed, create a script that takes a message as an argument and sends the message to the room. Then add it **notify** to etcd like this:

```
etcdctl set /goship/config '{"deploy_user":"YOUR_SSH_USER_ON_SERVER","notify":"/path/to/some/chat/notify.sh"}'
```

[Sevabot](http://sevabot-skype-bot.readthedocs.org/en/latest/) is a good choice for Skype.

# Tools

There are some tools added in the **/tools** directory that can be used interface with Goship

1) **goshipcfg**: It can be used to dump or restore etcd data as json. It can also be used to migrate from v1 config to current etcd data structure expected by Goship.

2) **deploy**:  Can be used as a script by the "deploy" to create a knife solo command which reads in the appropriate servers from ETCD and runs knife solo.

# Plugins

Goship suffices as a basic application to aid your deployments. However, you may wish to extend Goship with some custom UI on its home page with plugins.

To do so, head over to [Plugins](plugins).

GoShip was inspired by [Rackspace's Dreadnot](https://github.com/racker/dreadnot) ([UI image](http://c179631.r31.cf0.rackcdn.com/dreadnot-overview.png)) and [Etsy's Deployinator](https://github.com/etsy/deployinator/) ([UI image](http://farm5.staticflickr.com/4065/4620552264_9e0fdf634d_b.jpg)).

The GoShip logo is an adaptation of the [Go gopher](http://blog.golang.org/gopher) created by Renee French under the [Creative Commons Attribution 3.0 license](https://creativecommons.org/licenses/by/3.0/).

# Docker support (Experimental)
You can use Docker instead of Github to store target revision of deployment.

Configuration with docker looks like this:
   ```yaml
   deploy_user: YOUR_SSH_USER_ON_SERVER
   projects:
   - name: my-project
     repo_type: docker
     repo_owner: gcr.io
     repo_name: my-namespace/my-repo
     source:
       repo_name: my-project
       repo_owner: github-user-or-org
     envs:
     - name: staging
       deploy: "/tmp/deploy -p=my-project -e=staging"
       repo_path: ""
       hosts:
       - my-staging-server.example.com
       branch: latest
   ```

* `repo_owner` is used to specify docker registry.
  Currently it must be `gcr.io` (Google Container Registry)
* `repo_name` is used to specify docker image name
* `branch` in `envs` is used to specify docker image tag 
* You have to specify `source` section to keep corresponding github repository
* `repo_path` in `envs` is ignored
