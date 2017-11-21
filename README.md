# Docker Volume GIT

[![Go Report Card](https://goreportcard.com/badge/github.com/kassisol/docker-volume-git)](https://goreportcard.com/report/github.com/kassisol/docker-volume-git)

This plugin allows to mount git repository in container.

## Getting Started

### Install

```bash
$ docker plugin install kassisol/gitvol:x.x.x
```

### Create a volume

| Key           | Default   | Description |
|---------------|-----------|-------------|
| url           |           |             |
| ref           | master    |             |
| auth-type     | anonymous |             |
| auth-user     |           |             |
| secret-driver | stdin     |             |

#### As anonymous

```bash
$ docker volume create -d kassisol/gitvol:x.x.x -o "url=https://github.com/kassisol/docker-volume-git.git" vol_gitplugin
```

#### Using a secret on Standard Input

| Key                | Description |
|--------------------|-------------|
| auth-password      |             |

```bash
$ docker volume create -d kassisol/gitvol:x.x.x -o "url=ssh://<git_url>/<project>/<repo>" -o "auth-type=password" -o "auth-user=user1" -o "secret-driver=stdin" -o "auth-password=pass1234" vol_repo
```

#### Using a secret in Vault

| Key                | Description |
|--------------------|-------------|
| vault-addr         |             |
| vault-token        |             |
| vault-secret-path  |             |
| vault-secret-field |             |

```bash
$ docker volume create -d kassisol/gitvol:x.x.x -o "url=ssh://<git_url>/<project>/<repo>" -o "auth-type=pubkey" -o "auth-user=user1" -o "secret-driver=vault" -o "vault-addr=http://192.168.0.10:8200" -o "vault-token=1ad7bce4-078e-23a1-07e9-981a02abd514" -o "vault-secret-path=secret/user1" -o "vault-secret-field=prikey" vol_repo
```

## User Feedback

### Issues

If you have any problems with or questions about this application, please contact us through a [GitHub](https://github.com/kassisol/docker-volume-git/issues) issue.
